// ============================================
// SETTINGS
// ============================================

let allSettings = [];

async function loadSettingsData() {
    try {
        const data = await window.api.get_settings_data();
        allSettings = (data && data.categories) ? data.categories : [];
    } catch (e) {
        console.error('Settings load error:', e);
        allSettings = [];
    }
    const segCount = document.getElementById('seg-count-settings');
    if (segCount) segCount.textContent = allSettings.length;
}

function renderSettingsPane() {
    if (activeManageTab !== 'settings') return;
    renderSettingsList();
}

function renderSettingsList() {
    const list = document.getElementById('settings-list-manage');
    if (!list) return;

    const search = (document.getElementById('settings-search')?.value || '').toLowerCase();
    const filtered = allSettings.filter(s => {
        if (!search) return true;
        return (s.name || '').toLowerCase().includes(search) ||
               (s.folderPath || '').toLowerCase().includes(search);
    });

    if (!filtered.length) {
        list.innerHTML = `<div class="list-empty"><div class="list-empty-title">${t('empty_no_settings')}</div></div>`;
        return;
    }

    list.innerHTML = filtered.map((s, i) => {
        const idx = allSettings.indexOf(s);
        return `
            <div class="row" data-idx="${idx}">
                <div class="row-avatar">⚙️</div>
                <div class="row-main">
                    <div class="row-name">${escapeHtml(s.name || '')}</div>
                    <div class="row-meta">
                        <span>${escapeHtml(s.folderPath || '')}</span>
                    </div>
                </div>
                <div class="row-actions">
                    <button class="row-action" onclick="openSettingFolder(${idx})" title="${t('title_open_folder')}">${ICON.folder}</button>
                    <button class="row-action" onclick="copySettingPath(${idx})" title="${t('title_copy_path')}">${ICON.copy}</button>
                    <button class="row-action" onclick="editSetting(${idx})" title="${t('title_edit')}">${ICON.edit}</button>
                    <button class="row-action danger" onclick="deleteSetting(${idx})" title="${t('title_delete')}">${ICON.trash}</button>
                </div>
            </div>
        `;
    }).join('');
}

function openAddSetting() {
    const overlay = document.getElementById('setting-editor-overlay');
    if (!overlay) return;
    document.getElementById('setting-editor-title').textContent = t('setting_editor_add');
    document.getElementById('set-name').value = '';
    document.getElementById('set-folder').value = '';
    overlay.classList.add('active');
}

function editSetting(idx) {
    const s = allSettings[idx];
    if (!s) return;
    const overlay = document.getElementById('setting-editor-overlay');
    if (!overlay) return;
    document.getElementById('setting-editor-title').textContent = t('setting_editor_edit');
    document.getElementById('set-name').value = s.name || '';
    document.getElementById('set-folder').value = s.folderPath || '';
    overlay.dataset.editIdx = idx;
    overlay.classList.add('active');
}

function closeSettingEditor() {
    const overlay = document.getElementById('setting-editor-overlay');
    if (overlay) {
        overlay.classList.remove('active');
        delete overlay.dataset.editIdx;
    }
}

function openSettingFolder(idx) {
    const s = allSettings[idx];
    if (!s || !s.folderPath) return;
    window.api?.open_folder(s.folderPath);
}

async function copySettingPath(idx) {
    const s = allSettings[idx];
    if (!s || !s.folderPath) return;
    const ok = await copyToClipboard(s.folderPath);
    showToast(ok ? t('toast_path_copied_s') : t('err_copy_path'), !ok);
}

async function saveSetting() {
    const overlay = document.getElementById('setting-editor-overlay');
    if (!overlay) return;
    const name = document.getElementById('set-name').value.trim();
    const folder = document.getElementById('set-folder').value.trim();
    if (!name) { showToast(t('val_name_required'), true); return; }

    const editIdx = overlay.dataset.editIdx;
    if (editIdx !== undefined) {
        allSettings[parseInt(editIdx)] = { name, folderPath: folder };
    } else {
        allSettings.push({ name, folderPath: folder });
    }

    closeSettingEditor();
    await persistSettings();
    renderSettingsList();
}

async function deleteSetting(idx) {
    const s = allSettings[idx];
    if (!s) return;
    if (!(await uiConfirm(t('mng_confirm_delete_setting', { name: s.name || '' }), { danger: true }))) return;
    allSettings.splice(idx, 1);
    await persistSettings();
    renderSettingsList();
}

async function persistSettings() {
    try {
        const result = await window.api.save_settings_data(JSON.stringify({
            version: '1.0',
            categories: allSettings
        }));
        if (result && result.error) {
            showToast(t('mng_err_save', { error: tErr(result) }), true);
        } else {
            showToast(t('toast_saved'));
            const segCount = document.getElementById('seg-count-settings');
            if (segCount) segCount.textContent = allSettings.length;
        }
    } catch (e) {
        showToast(t('mng_err_save', { error: String(e) }), true);
    }
}
