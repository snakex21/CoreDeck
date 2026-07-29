// ============================================
// KATALOG — zarządzanie programami i sterownikami
// Linear-style: minimalna lista, akcje na hover
// ============================================

// --- Wspólny state ---
let activeManageTab = 'programs';
let selectedDriverType = null;

// --- Programs state ---
let allPrograms = [];

// --- Drivers state ---
let driversData = { version: "1.0", types: {} };
let driversSearchTerm = "";
let driverEditorMode = "series";
let driverEditContext = null;
let driversSearchTimer = null;

// --- SVG icons (inline, no external deps) ---
const ICON = {
    plus: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>',
    edit: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
    trash: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>',
    copy: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
    folder: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>',
    chevDown: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>',
    package: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="16.5" y1="9.4" x2="7.5" y2="4.21"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
    chip: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"/><rect x="9" y="9" width="6" height="6"/><line x1="9" y1="2" x2="9" y2="4"/><line x1="15" y1="2" x2="15" y2="4"/><line x1="9" y1="20" x2="9" y2="22"/><line x1="15" y1="20" x2="15" y2="22"/><line x1="20" y1="9" x2="22" y2="9"/><line x1="20" y1="15" x2="22" y2="15"/><line x1="2" y1="9" x2="4" y2="9"/><line x1="2" y1="15" x2="4" y2="15"/></svg>',
    inbox: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 16 12 14 15 10 15 8 12 2 12"/><path d="M5.45 5.11L2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z"/></svg>',
    clock: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>',
    archive: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="21 8 21 21 3 21 3 8"/><rect x="1" y="3" width="22" height="5"/><line x1="10" y1="12" x2="14" y2="12"/></svg>',
    download: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>',
};

// ============================================
// HELPERS
// ============================================

function setText(id, value) {
    const el = document.getElementById(id);
    if (el) el.textContent = value;
}

function getDriverStats() {
    const types = Object.values(driversData.types || {});
    let brands = 0;
    let series = 0;
    types.forEach(t => {
        const brandValues = Object.values(t.brands || {});
        brands += brandValues.length;
        brandValues.forEach(b => { series += (b.series || []).length; });
    });
    return { types: types.length, brands, series };
}

function updateManageStats() {
    const categories = new Set(allPrograms.map(p => p.category).filter(Boolean));
    const ds = getDriverStats();
    setText('stat-programs', allPrograms.length || 0);
    setText('stat-categories-sub', (categories.size || 0) + ' ' + t('stat_cats_suffix'));
    setText('stat-driver-types', ds.types || 0);
    setText('stat-driver-series-sub', (ds.series || 0) + ' ' + t('stat_series_suffix'));
    setText('seg-count-programs', allPrograms.length || 0);
    setText('seg-count-drivers', ds.types || 0);
    setText('seg-count-pages', allPrograms.filter(p => p.type === 'link').length || 0);
}

function escapeHtml(s) {
    if (s == null) return '';
    return String(s)
        .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

// Map a program installType (or the pseudo-type 'page') to a shared design
// system badge variant from style.css. Mirrors installBadgeClass() in
// configurator.js so the same catalog entry gets the same tag on both pages.
// Returns a fixed literal, so the caller never has to escape the result.
function badgeClass(installType) {
    switch (installType) {
        case 'installer': return 'badge badge-info';
        case 'portable':  return 'badge badge-success';
        case 'tool':
        case 'addon':
        case 'collection': return 'badge badge-purple';
        case 'page':      return 'badge badge-warn';
        default:          return 'badge badge-neutral';
    }
}

// ============================================
// CLIPBOARD + TOAST
// ============================================

async function copyToClipboard(text) {
    // Próba 1: nowoczesne API (działa w HTTPS i bezpiecznych kontekstach)
    if (navigator.clipboard && window.isSecureContext) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (e) {
            // spadamy do fallbacku
        }
    }

    // Fallback: textarea + execCommand('copy') — działa w webview bez uprawnień
    const ta = document.createElement('textarea');
    ta.value = text;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.top = '-1000px';
    ta.style.left = '-1000px';
    ta.style.opacity = '0';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    ta.setSelectionRange(0, text.length);
    let ok = false;
    try {
        ok = document.execCommand('copy');
    } catch (e) {
        ok = false;
    }
    document.body.removeChild(ta);
    return ok;
}

let toastTimer = null;
function showToast(message, isError) {
    let toast = document.getElementById('manage-toast');
    if (!toast) {
        toast = document.createElement('div');
        toast.id = 'manage-toast';
        document.body.appendChild(toast);
    }
    toast.textContent = message;
    toast.className = 'toast' + (isError ? ' toast-error' : '') + ' show';
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => {
        toast.classList.remove('show');
    }, 1800);
}

function defaultIconFor(installType) {
    if (installType === 'installer') return '⚙️';
    if (installType === 'tool') return '🧰';
    if (installType === 'addon') return '🧩';
    return '📦';
}

function isPathIcon(icon) {
    if (!icon) return false;
    // Traktuj jako ścieżkę jeśli ma separator lub rozszerzenie obrazka
    return /[\\/]/.test(icon) || /\.(png|jpe?g|svg|gif|ico|webp|bmp)$/i.test(icon);
}

function renderProgramIcon(p) {
    const fallback = defaultIconFor(p.installType);
    const icon = p.icon;
    if (!icon) return escapeHtml(fallback);
    if (isPathIcon(icon)) {
        // Renderuj jako <img>, a jeśli się nie załaduje — pokaż emoji
        const safePath = escapeHtml(icon);
        const safeFallback = escapeHtml(fallback);
        return `<img src="${safePath}" alt="" onerror="this.outerHTML='${safeFallback}'" style="width:18px;height:18px;object-fit:contain;display:block;" />`;
    }
    return escapeHtml(icon);
}

function renderIconPreview(iconValue) {
    const preview = document.getElementById('prog-icon-preview');
    const note = document.getElementById('prog-icon-note');
    if (!preview) return;
    const value = iconValue || '';
    if (!value) {
        preview.innerHTML = '📦';
        if (note) note.textContent = t('img_format_hint');
        return;
    }
    if (value.startsWith('data:image/') || isPathIcon(value)) {
        preview.innerHTML = `<img src="${escapeHtml(value)}" alt="" onerror="this.outerHTML='📦'" />`;
        if (note) note.textContent = value.startsWith('data:image/') ? t('icon_note_data_url') : value;
        return;
    }
    preview.textContent = value;
    if (note) note.textContent = t('icon_note_emoji');
}

function pickProgramIcon() {
    document.getElementById('prog-icon-file')?.click();
}

function clearProgramIcon() {
    const iconInput = document.getElementById('prog-icon');
    const fileInput = document.getElementById('prog-icon-file');
    if (iconInput) iconInput.value = '';
    if (fileInput) fileInput.value = '';
    renderIconPreview('');
}

function initProgramIconPicker() {
    const fileInput = document.getElementById('prog-icon-file');
    if (!fileInput || fileInput._wired) return;
    fileInput._wired = true;
    fileInput.addEventListener('change', () => {
        const file = fileInput.files && fileInput.files[0];
        if (!file) return;
        if (!file.type.startsWith('image/')) {
            showToast(t('error_not_image'), true);
            fileInput.value = '';
            return;
        }
        const reader = new FileReader();
        reader.onload = () => {
            const dataUrl = String(reader.result || '');
            document.getElementById('prog-icon').value = dataUrl;
            renderIconPreview(dataUrl);
            const note = document.getElementById('prog-icon-note');
            if (note) note.textContent = file.name + ' · zapisane jako obrazek w JSON';
        };
        reader.onerror = () => showToast(t('error_image_load'), true);
        reader.readAsDataURL(file);
    });
}

// ============================================
// THEME TOGGLE
// ============================================

function initThemeToggle() {
    const btn = document.getElementById('theme-toggle-btn');
    if (!btn) return;
    const iconSlot = btn.querySelector('.theme-icon');
    const isLight = localStorage.getItem('launcher-theme') === 'light';
    if (isLight) {
        document.body.classList.add('light-theme');
        if (iconSlot) iconSlot.textContent = '☀️';
    }
    btn.addEventListener('click', () => {
        document.body.classList.toggle('light-theme');
        const nowLight = document.body.classList.contains('light-theme');
        localStorage.setItem('launcher-theme', nowLight ? 'light' : 'dark');
        if (iconSlot) iconSlot.textContent = nowLight ? '☀️' : '🌙';
    });
}

// ============================================
// TABS (segmented control)
// ============================================

function initManageTabs() {
    document.querySelectorAll('.seg').forEach(tab => {
        tab.addEventListener('click', () => {
            const target = tab.dataset.tab;
            if (target === activeManageTab) return;

            document.querySelectorAll('.seg').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            document.querySelectorAll('.tab-pane').forEach(c => c.classList.remove('active'));
            document.getElementById('pane-' + target).classList.add('active');

            document.getElementById('toolbar-programs').style.display = target === 'programs' ? 'flex' : 'none';
            document.getElementById('toolbar-drivers').style.display = target === 'drivers' ? 'flex' : 'none';
            document.getElementById('toolbar-settings').style.display = target === 'settings' ? 'flex' : 'none';
            document.getElementById('toolbar-knowledge').style.display = target === 'knowledge' ? 'flex' : 'none';
            document.getElementById('toolbar-pages').style.display = target === 'pages' ? 'flex' : 'none';

            activeManageTab = target;
            if (target === 'programs') renderPrograms();
            else if (target === 'drivers') renderDriversPane();
            else if (target === 'settings') renderSettingsPane();
            else if (target === 'knowledge') renderKnowledgePane();
            else if (target === 'pages') renderPagesPane();
            else if (target === 'import-export') renderImportExportPane();
        });
    });
}
