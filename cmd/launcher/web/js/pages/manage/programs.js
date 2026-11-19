// ============================================
// PROGRAMS
// ============================================

async function loadPrograms() {
    try {
        const data = await window.api.get_programs_data();
        if (data.error) throw new Error(data.error);
        allPrograms = data.programs || [];
    } catch (e) {
        console.error('Failed to load programs:', e);
        allPrograms = [];
    }
    updateManageStats();
    renderPrograms();
    populateCategories();
}

function renderPrograms() {
    if (activeManageTab !== 'programs') return;
    const list = document.getElementById('programs-list');
    if (!list) return;

    const search = (document.getElementById('search')?.value || '').toLowerCase();
    const filterType = document.getElementById('filter-type')?.value;
    const filterCat = document.getElementById('filter-category')?.value;

    const filtered = allPrograms.filter(p => {
        if (search) {
            const hay = [p.name, p.description, ...(p.tags || [])].join(' ').toLowerCase();
            if (!hay.includes(search)) return false;
        }
        if (filterType && filterType !== 'all' && p.installType !== filterType) return false;
        if (filterCat && filterCat !== 'all' && p.category !== filterCat) return false;
        return true;
    });

    if (filtered.length === 0) {
        list.innerHTML = `
            <div class="list-empty">
                ${ICON.inbox}
                <div class="list-empty-title">${search ? t('empty_no_matches') : t('empty_no_programs')}</div>
                <div>${search ? t('empty_try_other') : t('empty_add_first')}</div>
            </div>
        `;
        return;
    }

    list.innerHTML = filtered.map((p) => {
        const idx = allPrograms.indexOf(p);
        const nameAttr = encodeURIComponent(p.name);
        const installType = p.installType || 'other';
        const cat = p.category || '—';

        let versionBadge = '';
        if (p.variants && p.variants.length > 0) {
            const variant = p.variants.find((v, vi) => vi === p.defaultVariant) || p.variants[0];
            versionBadge = `<span class="row-version" title="${t('badge_variant')}">${escapeHtml(variant.name || t('mng_variant_default'))} ${ICON.chevDown}</span>`;
        }

        const isPage = p.type === 'link';
        const tagLabel = isPage ? t('chip_page_label') : escapeHtml(installType);
        const tagClass = badgeClass(isPage ? 'page' : installType);

        return `
            <div class="row">
                <div class="row-avatar">${renderProgramIcon(p)}</div>
                <div class="row-main">
                    <div class="row-name">${escapeHtml(p.name)}</div>
                    <div class="row-meta">
                        <span>${escapeHtml(cat)}</span>
                        <span class="row-meta-sep"></span>
                        <span>${escapeHtml(p.description || t('no_description'))}</span>
                    </div>
                </div>
                <span class="${tagClass}">${tagLabel}</span>
                <div class="row-actions">
                    ${versionBadge}
                    <button class="row-action" onclick="openEditProgram('${nameAttr}')" title="${t('title_edit')}">${ICON.edit}</button>
                    <button class="row-action danger" onclick="deleteProgramUI('${nameAttr}')" title="${t('title_delete')}">${ICON.trash}</button>
                </div>
            </div>
        `;
    }).join('');

    updateManageStats();
}

function populateCategories() {
    const select = document.getElementById('filter-category');
    if (!select) return;
    const cats = [...new Set(allPrograms.map(p => p.category).filter(Boolean))].sort();
    select.innerHTML = `<option value="all">${escapeHtml(t('filter_all_cats'))}</option>`;
    cats.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c;
        opt.textContent = c;
        select.appendChild(opt);
    });
}

// --- Program editor ---

function fillProgramForm(p) {
    document.getElementById('prog-name').value = p ? p.name : '';
    document.getElementById('prog-name').disabled = !!p;
    document.getElementById('prog-desc').value = p ? (p.description || '') : '';
    document.getElementById('prog-category').value = p ? (p.category || '') : '';
    document.getElementById('prog-icon').value = p ? (p.icon || '') : '';
    renderIconPreview(p ? (p.icon || '') : '');
    document.getElementById('prog-type').value = p ? (p.type || 'folder') : 'folder';
    document.getElementById('prog-install-type').value = p ? (p.installType || 'portable') : 'portable';
    document.getElementById('prog-folder').value = p ? (p.folderPath || '') : '';
    document.getElementById('prog-installer-file').value = p ? (p.installerFile || '') : '';
    document.getElementById('prog-installer-args').value = p ? (p.installerArgs || '') : '';
    document.getElementById('prog-run-from-temp').checked = p ? !!p.runFromTemp : false;
    document.getElementById('prog-localpath').value = p ? (p.localPath || '') : '';
    document.getElementById('prog-externalurl').value = p ? (p.externalUrl || '') : '';
    document.getElementById('prog-details').value = p ? (p.details || '') : '';
    const linksList = document.getElementById('prog-links-list');
    if (linksList) {
        linksList.innerHTML = '';
        const links = p ? (p.links || []) : [];
        links.forEach(link => addProgramLinkRow(link.label, link.url, link.kind));
    }
    document.getElementById('prog-tags').value = p ? (p.tags || []).join(', ') : '';
    const relevant = p ? (p.profileRelevant || []) : [];
    document.querySelectorAll('.prog-profile-cb').forEach(cb => {
        cb.checked = relevant.includes(cb.value);
    });
    const versionsList = document.getElementById('prog-versions-list');
    if (versionsList) {
        versionsList.innerHTML = '';
        const variants = p ? (p.variants || []) : [];
        variants.forEach(v => addVersionRow(v.name, v.folderPath, v.installType, v.installerFile, v.installerArgs, !!v.runFromTemp));
    }
}

function addProgramLinkRow(label = '', url = '', kind = 'guide') {
    const list = document.getElementById('prog-links-list');
    if (!list) return;
    const row = document.createElement('div');
    row.className = 'program-link-row';
    const kinds = [
        ['guide', t('mng_link_kind_guide')],
        ['docs', t('mng_link_kind_docs')],
        ['download', t('mng_link_kind_download')],
        ['official', t('mng_link_kind_official')],
        ['other', t('mng_link_kind_other')]
    ];
    row.innerHTML = `
        <input type="text" class="form-input program-link-label" placeholder="${escapeHtml(t('mng_ph_link_label'))}" value="${escapeHtml(label || '')}" />
        <input type="text" class="form-input program-link-url" placeholder="${escapeHtml(t('mng_ph_link_url'))}" value="${escapeHtml(url || '')}" />
        <select class="form-select program-link-kind">
            ${kinds.map(([value, text]) => `<option value="${value}" ${value === (kind || 'guide') ? 'selected' : ''}>${escapeHtml(text)}</option>`).join('')}
        </select>
        <button type="button" title="${t('title_delete')}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
    `;
    row.querySelector('button').addEventListener('click', () => row.remove());
    list.appendChild(row);
}

function addVersionRow(label, folderPath, installType, installerFile, installerArgs, runFromTemp) {
    const list = document.getElementById('prog-versions-list');
    if (!list) return;
    const row = document.createElement('div');
    row.className = 'version-row';
    const types = ['portable', 'installer', 'script', 'tool', 'addon', 'other'];
    const selType = installType || 'portable';
    const options = types.map(ty => `<option value="${ty}" ${ty === selType ? 'selected' : ''}>${ty}</option>`).join('');
    row.innerHTML = `
        <input type="text" class="form-input" placeholder="${escapeHtml(t('ph_variant_label'))}"value="${label ? String(label).replace(/"/g, '&quot;') : ''}" />
        <input type="text" class="form-input" placeholder="folderPath" value="${folderPath ? String(folderPath).replace(/"/g, '&quot;') : ''}" />
        <input type="text" class="form-input" placeholder="installerFile" value="${escapeHtml(installerFile || '')}" />
        <input type="text" class="form-input" placeholder="args" value="${escapeHtml(installerArgs || '')}" />
        <label class="form-check" style="margin:0;"><input type="checkbox" class="version-temp" ${runFromTemp ? 'checked' : ''} /> TEMP</label>
        <select class="form-select">${options}</select>
        <button type="button" title="${t('title_delete')}"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg></button>
    `;
    row.querySelector('button').addEventListener('click', () => row.remove());
    list.appendChild(row);
}

function openAddProgram() {
    document.getElementById('program-editor-title').textContent = t('editor_add_program');
    fillProgramForm(null);
    document.getElementById('program-editor-overlay').classList.add('active');
    setTimeout(() => document.getElementById('prog-name')?.focus(), 50);
}

function openEditProgram(encodedName) {
    const name = decodeURIComponent(encodedName);
    const p = allPrograms.find(x => x.name === name);
    if (!p) return;
    document.getElementById('program-editor-title').textContent = t('mng_editor_edit_program');
    fillProgramForm(p);
    document.getElementById('program-editor-overlay').classList.add('active');
}

function closeProgramEditor() {
    document.getElementById('program-editor-overlay').classList.remove('active');
}

async function saveNewProgram() {
    const name = document.getElementById('prog-name').value.trim();
    if (!name) { uiToast(t('err_name_required'), { kind: 'error' }); return; }

    const tags = document.getElementById('prog-tags').value.split(',').map(t => t.trim()).filter(t => t);
    const profileRelevant = Array.from(document.querySelectorAll('.prog-profile-cb'))
        .filter(cb => cb.checked).map(cb => cb.value);

    const variants = [];
    document.querySelectorAll('#prog-versions-list .version-row').forEach(row => {
        const inputs = row.querySelectorAll('input');
        const vName = inputs[0].value.trim();
        const folderPath = inputs[1].value.trim();
        const installerFile = inputs[2].value.trim();
        const installerArgs = inputs[3].value.trim();
        const runFromTemp = !!row.querySelector('.version-temp')?.checked;
        const installType = row.querySelector('select').value;
        if (!vName && !folderPath) return;
        const variant = { name: vName, folderPath, installType };
        if (installerFile) variant.installerFile = installerFile;
        if (installerArgs) variant.installerArgs = installerArgs;
        if (runFromTemp) variant.runFromTemp = true;
        variants.push(variant);
    });

    const links = [];
    document.querySelectorAll('#prog-links-list .program-link-row').forEach(row => {
        const label = row.querySelector('.program-link-label')?.value.trim() || '';
        const url = row.querySelector('.program-link-url')?.value.trim() || '';
        const kind = row.querySelector('.program-link-kind')?.value || 'other';
        if (!url) return;
        links.push({ label: label || url, url, kind });
    });

    const obj = {
        name,
        description: document.getElementById('prog-desc').value.trim(),
        category: document.getElementById('prog-category').value.trim(),
        icon: document.getElementById('prog-icon').value.trim(),
        type: document.getElementById('prog-type').value,
        folderPath: document.getElementById('prog-folder').value.trim(),
        localPath: document.getElementById('prog-localpath').value.trim(),
        externalUrl: document.getElementById('prog-externalurl').value.trim(),
        details: document.getElementById('prog-details').value,
        links,
        installType: document.getElementById('prog-install-type').value,
        tags,
        profileRelevant
    };
    const installerFile = document.getElementById('prog-installer-file').value.trim();
    const installerArgs = document.getElementById('prog-installer-args').value.trim();
    if (installerFile) obj.installerFile = installerFile;
    if (installerArgs) obj.installerArgs = installerArgs;
    if (document.getElementById('prog-run-from-temp').checked) obj.runFromTemp = true;
    if (variants.length > 0) {
        obj.variants = variants;
        obj.defaultVariant = 0;
    }

    try {
        const result = await window.api.add_program(JSON.stringify(obj));
        if (result.error) { uiToast(t('mng_err_generic', { error: tErr(result) }), { kind: 'error' }); return; }
        closeProgramEditor();
        await loadPrograms();
    } catch (e) {
        uiToast(t('mng_err_save', { error: String(e) }), { kind: 'error' });
    }
}

async function deleteProgramUI(encodedName) {
    const name = decodeURIComponent(encodedName);
    if (!(await uiConfirm(t('mng_confirm_delete_program', { name }), { danger: true }))) return;
    try {
        const result = await window.api.delete_program(name);
        if (result.error) { uiToast(t('mng_err_generic', { error: tErr(result) }), { kind: 'error' }); return; }
        await loadPrograms();
    } catch (e) {
        uiToast(t('mng_err_delete', { error: String(e) }), { kind: 'error' });
    }
}

function initProgramsSearch() {
    document.getElementById('search')?.addEventListener('input', renderPrograms);
    document.getElementById('filter-type')?.addEventListener('change', renderPrograms);
    document.getElementById('filter-category')?.addEventListener('change', renderPrograms);
}

// --- Install moved to 1 Konfiguracja ---
