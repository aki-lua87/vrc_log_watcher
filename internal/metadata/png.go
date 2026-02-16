package metadata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"vrc_log_watcher/internal/applog"
)

// ReadPNG PNGファイルからVRChatのメタデータを読み取る
func ReadPNG(filePath string, logger applog.Logger) map[string]string {
	metadata := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		logger.Log(fmt.Sprintf("[PNG Metadata] 画像ファイルを開けませんでした: %v", err))
		return metadata
	}
	defer file.Close()

	// PNGシグネチャを確認（8バイト）
	signature := make([]byte, 8)
	if _, err := file.Read(signature); err != nil {
		logger.Log(fmt.Sprintf("[PNG Metadata] PNGシグネチャの読み取りに失敗: %v", err))
		return metadata
	}

	expectedSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.Equal(signature, expectedSignature) {
		logger.Log("[PNG Metadata] 有効なPNGファイルではありません")
		return metadata
	}

	logger.Log(fmt.Sprintf("[PNG Metadata] PNGファイルの解析を開始: %s", filePath))

	chunkCount := 0
	for {
		lengthBytes := make([]byte, 4)
		if _, err := file.Read(lengthBytes); err != nil {
			break
		}
		length := binary.BigEndian.Uint32(lengthBytes)

		chunkType := make([]byte, 4)
		if _, err := file.Read(chunkType); err != nil {
			break
		}

		chunkTypeStr := string(chunkType)
		chunkCount++

		chunkData := make([]byte, length)
		if _, err := file.Read(chunkData); err != nil {
			break
		}

		// CRC（4バイト）をスキップ
		crc := make([]byte, 4)
		if _, err := file.Read(crc); err != nil {
			break
		}

		if chunkTypeStr == "tEXt" {
			parseTEXt(chunkData, metadata, logger)
		}

		if chunkTypeStr == "iTXt" {
			parseITXt(chunkData, metadata, logger)
		}

		if chunkTypeStr == "zTXt" {
			parseZTXt(chunkData, logger)
		}

		if chunkTypeStr == "IEND" {
			logger.Log(fmt.Sprintf("[PNG Metadata] IENDチャンクに到達。合計 %d チャンクを処理", chunkCount))
			break
		}
	}

	logger.Log(fmt.Sprintf("[PNG Metadata] 抽出されたメタデータ: %d 件", len(metadata)))
	for k, v := range metadata {
		logger.Log(fmt.Sprintf("[PNG Metadata]   %s: %s", k, v))
	}

	extractXMPMetadata(metadata, logger)

	return metadata
}

func parseTEXt(data []byte, metadata map[string]string, logger applog.Logger) {
	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex > 0 && nullIndex < len(data)-1 {
		key := string(data[:nullIndex])
		value := string(data[nullIndex+1:])
		metadata[key] = value
		logger.Log(fmt.Sprintf("[PNG Metadata] tEXt: %s = %s", key, value))
	}
}

func parseITXt(data []byte, metadata map[string]string, logger applog.Logger) {
	if len(data) == 0 {
		return
	}

	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex <= 0 || nullIndex >= len(data)-1 {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（keyword）")
		return
	}

	key := string(data[:nullIndex])
	pos := nullIndex + 1

	if pos >= len(data) {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（compression flag）")
		return
	}
	compressionFlag := data[pos]
	pos++

	if pos >= len(data) {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（compression method）")
		return
	}
	pos++ // compression methodをスキップ

	nullIndex2 := bytes.IndexByte(data[pos:], 0)
	if nullIndex2 < 0 {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（language tag）")
		return
	}
	languageTag := string(data[pos : pos+nullIndex2])
	pos += nullIndex2 + 1

	if pos >= len(data) {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（translated keyword）")
		return
	}
	nullIndex3 := bytes.IndexByte(data[pos:], 0)
	if nullIndex3 < 0 {
		logger.Log("[PNG Metadata] iTXt: 不正なフォーマット（translated keyword終端）")
		return
	}
	translatedKeyword := string(data[pos : pos+nullIndex3])
	pos += nullIndex3 + 1

	if pos < len(data) {
		value := string(data[pos:])
		metadata[key] = value
		logger.Log(fmt.Sprintf("[PNG Metadata] iTXt: key=%s, lang=%s, translated=%s, value=%s, compressed=%d",
			key, languageTag, translatedKeyword, value, compressionFlag))
	}
}

func parseZTXt(data []byte, logger applog.Logger) {
	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex > 0 && nullIndex+2 < len(data) {
		key := string(data[:nullIndex])
		logger.Log(fmt.Sprintf("[PNG Metadata] zTXt チャンクを検出（スキップ）: %s", key))
	}
}

// extractXMPMetadata XMPメタデータからVRChat情報を抽出
func extractXMPMetadata(metadata map[string]string, logger applog.Logger) {
	xmpData, ok := metadata["XML:com.adobe.xmp"]
	if !ok {
		return
	}

	if startIdx := strings.Index(xmpData, "<vrc:WorldDisplayName>"); startIdx >= 0 {
		startIdx += len("<vrc:WorldDisplayName>")
		if endIdx := strings.Index(xmpData[startIdx:], "</vrc:WorldDisplayName>"); endIdx >= 0 {
			worldName := xmpData[startIdx : startIdx+endIdx]
			metadata["World Display Name"] = worldName
			logger.Log(fmt.Sprintf("[PNG Metadata] XMPから抽出: World Display Name = %s", worldName))
		}
	}

	if startIdx := strings.Index(xmpData, "<vrc:WorldID>"); startIdx >= 0 {
		startIdx += len("<vrc:WorldID>")
		if endIdx := strings.Index(xmpData[startIdx:], "</vrc:WorldID>"); endIdx >= 0 {
			worldID := xmpData[startIdx : startIdx+endIdx]
			metadata["World ID"] = worldID
			logger.Log(fmt.Sprintf("[PNG Metadata] XMPから抽出: World ID = %s", worldID))
		}
	}
}
