<script lang="ts">
    // アプリのロゴ（案B「分岐」）。テーマに関係なくティールで固定
    export let size = 32;
</script>

<span class="logo" style="width:{size}px;height:{size}px">
    <svg viewBox="0 0 64 64" width={size * 0.66} height={size * 0.66} aria-hidden="true">
        <g fill="none" stroke="currentColor" stroke-width="6" stroke-linecap="round" stroke-linejoin="round">
            <path d="M10 20h12M10 32h20M10 44h10" />
            <path d="M30 32h4c6 0 6-12 12-12h3M34 32h15M34 32c6 0 6 12 12 12h3" />
            <circle cx="53" cy="20" r="4" fill="currentColor" stroke="none" />
            <circle cx="53" cy="32" r="4" fill="currentColor" stroke="none" />
            <circle cx="53" cy="44" r="4" fill="currentColor" stroke="none" />
        </g>
    </svg>
</span>

<style>
    .logo {
        border-radius: 8px;
        background: #0d9488;
        color: #fff;
        display: grid;
        place-items: center;
        flex: none;
    }
</style>
