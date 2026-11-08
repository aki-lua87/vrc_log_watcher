VRC LogBridge
==============================

VRChat のログをリアルタイムに読み取り、ワールドへの入室、ユーザーの入退室、
スクリーンショットの撮影などを検知して、XSOverlay・Discord・Web などへ通知する
Windows 用のアプリです。（旧名: VRC Log Watcher）


■ 使い方
------------------------------
1. この zip を好きなフォルダに展開します。
2. VRCLogBridge.exe を起動します。
3. 初回起動時に、VRChat のログフォルダを自動で探して設定します。

※ 起動時に「Windows によって PC が保護されました」と表示された場合は、
   「詳細情報」→「実行」の順に押すと起動できます。

※ 起動しない場合は、Microsoft Edge WebView2 Runtime をインストールしてください。
   https://developer.microsoft.com/microsoft-edge/webview2/


■ 設定ファイル
------------------------------
設定は VRCLogBridge.exe と同じフォルダの setting.json に保存されます。

旧バージョン（vlw.exe）から移行する場合は、vlw.exe があったフォルダに
VRCLogBridge.exe を置くと、今までの設定をそのまま使えます。

更新するときも、同じフォルダの VRCLogBridge.exe を上書きしてください。


■ 配布元・最新版
------------------------------
https://github.com/aki-lua87/vrc_log_watcher/releases


■ ライセンス
------------------------------
本アプリは MIT License です（LICENSE を参照）。
アプリ内で使用しているフォント Nunito は SIL Open Font License 1.1 です（OFL.txt を参照）。

作者: aki_lua87
