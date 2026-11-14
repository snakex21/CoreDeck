let allPrograms = [];
let allTweaks = [];
let profiles = {};
let customProfiles = [];
let systemInfo = null;
let tweakStatesScanned = false;
let tweakStatesScanning = false;
let showRemovedTweaks = false;
let hideActiveTweaks = false;
let hideRemovedDebloat = false;
let currentActiveProfile = null;
let currentActiveProfileIsCustom = false;
let tweakEditorProfileId = null;
let tweakEditorSelectedTweaks = [];
let tweakEditorMode = 'tweaks';

function plur(n, one, few, many) {
    if (n === 1) return `${n} ${one}`;
    if (n % 10 >= 2 && n % 10 <= 4 && (n % 100 < 10 || n % 100 >= 20)) return `${n} ${few}`;
    return `${n} ${many}`;
}

function escapeHtml(s) {
    return String(s || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function renderProgressResult(r) {
    const icon = r.success ? '✅' : '❌';
    const cls = r.success ? 'ok' : 'err';
    const pathBtn = r.path
        ? ` <button class="btn-mini progress-open-folder" data-path="${escapeHtml(r.path)}">📂 Otwórz folder</button>`
        : '';
    return `<div class="${cls}">${icon} ${escapeHtml(r.name)} — ${escapeHtml(r.message)}${pathBtn}</div>`;
}
let editingProfileId = null;
let editorSelectedPrograms = []; // global state for profile editor
let editorSelectedTweaks = []; // global state for profile editor tweaks

document.addEventListener('DOMContentLoaded', async () => {
    initThemeToggle();
    initTabs();
    // Load all data in parallel — saves several round-trips
    await Promise.allSettled([
        loadSystemInfo(),
        loadPrograms(),
        loadCustomProfiles(),
        loadTweaks(),
    ]);
    // Defer tweak state scanning — it's heavy (100 checks).
    // Run it in background so the UI stays responsive.
    setTimeout(() => scanTweakStates(false), 100);
    initSearch();
    initProgressLogActions();
    loadTargetDir();
});

function initProgressLogActions() {
    const log = document.getElementById('progress-log');
    if (!log) return;
    log.addEventListener('click', async (e) => {
        const btn = e.target.closest('.progress-open-folder');
        if (!btn) return;
        const path = btn.dataset.path;
        if (path && window.api && typeof window.api.open_folder === 'function') {
            await window.api.open_folder(path);
        }
    });
}

// --- Theme Toggle ---
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

// --- Tabs ---
function initTabs() {
    document.querySelectorAll('.config-tab').forEach(tab => {
        tab.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                tab.click();
            }
        });
        tab.addEventListener('click', () => {
            document.querySelectorAll('.config-tab').forEach(t => {
                t.classList.remove('active');
                t.setAttribute('aria-selected', 'false');
            });
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            tab.classList.add('active');
            tab.setAttribute('aria-selected', 'true');
            document.getElementById('tab-' + tab.dataset.tab).classList.add('active');
            if (tab.dataset.tab === 'tweaki') {
                scanTweakStates(false);
            }
            if (tab.dataset.tab === 'debloat') {
                scanTweakStates(false);
                renderDebloat();
            }
        });
    });
}

// --- System Info ---
async function loadSystemInfo() {
    try {
        const info = await window.api.get_system_info();
        systemInfo = info;
        const el = document.getElementById('sys-info');
        el.innerHTML = `
            <div class="sys-info-card" style="margin-bottom:12px;">
                <div class="label">${t('sys_os')}</div>
                <div class="value small">${info.os || 'Unknown'}</div>
            </div>
            <div class="sys-info-card" style="margin-bottom:12px;">
                <div class="label">${t('sys_cpu')}</div>
                <div class="value small">${info.cpu || 'Unknown'}</div>
            </div>
            <div class="sys-info-card" style="margin-bottom:12px;">
                <div class="label">${t('sys_gpu')}</div>
                <div class="value small">${(info.gpu || []).join(', ') || 'Unknown'}</div>
            </div>
            <div class="sys-info-card" style="margin-bottom:12px;">
                <div class="label">${t('sys_ram')}</div>
                <div class="value">${info.ram || '?'}</div>
            </div>
            <div class="sys-info-card" style="margin-bottom:12px;">
                <div class="label">${t('sys_disk')}</div>
                <div class="value">${info.diskType || '?'}</div>
            </div>
            <div class="sys-info-card">
                <div class="label">${t('sys_type')}</div>
                <div class="value">${info.isLaptop ? t('sys_laptop') : t('sys_desktop')}</div>
            </div>
        `;
    } catch (e) {
        document.getElementById('sys-info').innerHTML = `<p style="color:var(--danger)">${t('err_sys_info')}</p>`;
    }
}

// --- Programs ---
async function loadPrograms() {
    try {
        const data = await window.api.get_programs_data();
        if (data.error) {
            throw new Error(data.error);
        }
        allPrograms = data.programs || [];
        profiles = data.profiles || {};
        renderProfiles();
        renderPrograms();
        populateCategories();
    } catch (e) {
        console.error('Failed to load programs:', e);
        const list = document.getElementById('programs-list');
        if (list) list.innerHTML = `<div class="program-item"><div class="info"><div class="name" style="color:var(--danger)">${t('err_programs_load')}</div><div class="desc">${String(e)}</div></div></div>`;
    }
}
