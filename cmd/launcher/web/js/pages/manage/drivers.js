// ============================================
// DRIVERS
// ============================================

async function loadDrivers() {
    try {
        const dd = await window.api.get_drivers_data();
        if (dd && dd.types) driversData = dd;
        else driversData = { version: "1.0", types: {} };
    } catch (e) {
        console.error('Nie udało się załadować sterowników:', e);
        driversData = { version: "1.0", types: {} };
    }
    updateManageStats();
    const search = document.getElementById('drivers-search');
    if (search && !search._wired) {
        search._wired = true;
        search.addEventListener('input', (e) => {
            driversSearchTerm = e.target.value.toLowerCase().trim();
            clearTimeout(driversSearchTimer);
            driversSearchTimer = setTimeout(renderDriversPane, 80);
        });
    }
    renderDriversPane();
}

function saveDrivers() {
    return window.api.save_drivers_data(JSON.stringify(driversData))
        .then((r) => {
            if (r && r.error) { alert(t('err_save_prefix') + r.error); return false; }
            return true;
        })
        .catch((e) => { alert(t('err_save_prefix') + String(e)); return false; });
}

function sortedTypeEntries() {
    return Object.entries(driversData.types || {})
        .map(([key, t]) => [key, t.title || key])
        .sort((a, b) => a[1].localeCompare(b[1], 'pl'));
}

function sortedBrandEntries(typeKey) {
    const t = driversData.types[typeKey];
    if (!t || !t.brands) return [];
    return Object.entries(t.brands)
        .map(([key, b]) => [key, b.title || key])
        .sort((a, b) => a[1].localeCompare(b[1], 'pl'));
}

function renderDriversPane() {
    updateManageStats();
    if (activeManageTab !== 'drivers') return;
    const list = document.getElementById('drivers-types-list');
    const detail = document.getElementById('drivers-detail');
    if (!list || !detail) return;

    const types = sortedTypeEntries();
    const term = driversSearchTerm;
    const filtered = types.filter(([typeKey, typeTitle]) => {
        if (!term) return true;
        if (typeTitle.toLowerCase().includes(term)) return true;
        const t = driversData.types[typeKey];
        for (const [brandKey, brand] of Object.entries(t.brands || {})) {
            if ((brand.title || brandKey).toLowerCase().includes(term)) return true;
            for (const s of (brand.series || [])) {
                if ((s.name || '').toLowerCase().includes(term)) return true;
            }
        }
        return false;
    });

    if (filtered.length === 0) {
        list.innerHTML = `<div style="padding:24px 18px;color:var(--text-muted);font-size:var(--fs-xs);">${term ? t('empty_no_matches') : t('empty_no_driver_types')}</div>`;
        detail.innerHTML = `
            <div class="drv-empty-detail">
                ${ICON.chip}
                <div class="drv-empty-detail-title">${term ? 'Brak wyników' : t('empty_select_type')}</div>
                <div>${term ? t('empty_try_other_drivers') : t('empty_add_first_type')}</div>
            </div>
        `;
        return;
    }

    // If selected type not in filtered, reset
    if (!filtered.find(([k]) => k === selectedDriverType)) {
        selectedDriverType = filtered[0][0];
    }

    list.innerHTML = filtered.map(([typeKey, typeTitle]) => {
        const t = driversData.types[typeKey];
        const brandCount = Object.keys(t.brands || {}).length;
        let seriesCount = 0;
        Object.values(t.brands || {}).forEach(b => { seriesCount += (b.series || []).length; });
        const active = typeKey === selectedDriverType;
        return `
            <button class="drv-type ${active ? 'active' : ''}" data-type="${escapeHtml(typeKey)}">
                <span class="drv-type-icon">${ICON.chip}</span>
                <span class="drv-type-name">${escapeHtml(typeTitle)}</span>
                <span class="drv-type-count">${brandCount} · ${seriesCount}</span>
            </button>
        `;
    }).join('');

    if (!list._wired) {
        list._wired = true;
        list.addEventListener('click', (e) => {
            const btn = e.target.closest('.drv-type');
            if (!btn) return;
            selectedDriverType = btn.dataset.type;
            renderDriversPane();
        });
    }

    renderDriverDetail(selectedDriverType);
}

function renderDriverDetail(typeKey) {
    const detail = document.getElementById('drivers-detail');
    if (!detail || !typeKey) return;
    // UWAGA: nie nazywaj tej zmiennej `t` — przysłoniłaby globalną funkcję
    // tłumaczeń t() używaną niżej w tym samym szablonie.
    const typeData = driversData.types[typeKey];
    if (!typeData) {
        detail.innerHTML = '';
        return;
    }

    const brands = sortedBrandEntries(typeKey);
    const totalSeries = brands.reduce((s, [bk]) => s + ((typeData.brands[bk].series || []).length), 0);

    detail.innerHTML = `
        <div class="drv-head">
            <div class="drv-head-main">
                <h2>${escapeHtml(typeData.title || typeKey)}</h2>
                <p>${brands.length} producent${brands.length === 1 ? '' : 'ów'} · ${totalSeries} ${t('unit_series')}</p>
            </div>
            <!-- btn-sm: four secondary actions hung off a section heading is
                 the densest cluster on the page; the same pattern in
                 configurator.html uses the compact size too. -->
            <div class="drv-head-actions">
                <button class="btn btn-secondary btn-sm" data-act="add-brand" title="${t('title_add_brand')}">${ICON.plus} Producent</button>
                <button class="btn btn-secondary btn-sm" data-act="add-series" title="${t('title_add_series')}">${ICON.plus} Seria</button>
                <button class="btn btn-secondary btn-sm" data-act="edit-type" title="${t('title_edit_type')}">${ICON.edit} Edytuj</button>
                <button class="btn btn-danger btn-sm" data-act="del-type" title="${t('title_delete_type')}">${ICON.trash} Usuń</button>
            </div>
        </div>
        ${brands.length === 0 ? `
            <div class="drv-empty-detail" style="height:auto;padding:60px 20px;">
                ${ICON.package}
                <div class="drv-empty-detail-title">Brak producentów</div>
                <div>${t('empty_add_first_brand')}</div>
            </div>
        ` : brands.map(([brandKey, brandTitle]) => {
            const brand = typeData.brands[brandKey];
            const series = brand.series || [];
            return `
                <div class="drv-brand-group">
                    <div class="drv-brand-head">
                        <div class="drv-brand-title">${escapeHtml(brandTitle)} <small>${series.length} ${t('unit_series')}</small></div>
                        <div class="drv-brand-actions">
                            <button data-act="add-series" data-brand="${escapeHtml(brandKey)}" title="${t('title_add_series_short')}">+</button>
                            <button data-act="edit-brand" data-brand="${escapeHtml(brandKey)}" title="${t('title_edit_brand')}">${ICON.edit}</button>
                            <button class="danger" data-act="del-brand" data-brand="${escapeHtml(brandKey)}" title="${t('title_delete_brand')}">${ICON.trash}</button>
                        </div>
                    </div>
                    ${series.length === 0 ? `
                        <div style="font-size:var(--fs-xs);color:var(--text-muted);padding:6px 0 12px;">${t('empty_no_series')}</div>
                    ` : `
                        <div class="series-list">
                            ${series.map((s, idx) => `
                                <div class="series-row">
                                    <div style="min-width:0;">
                                        <div class="series-name">${escapeHtml(s.name)}</div>
                                        <div class="series-path" title="${escapeHtml(s.folderPath)}">${escapeHtml(s.folderPath)}</div>
                                    </div>
                                    <div class="series-actions">
                                        <button data-act="copy-path" data-path="${escapeHtml(s.folderPath)}" title="${t('title_copy_path')}">${ICON.copy}</button>
                                        <button data-act="open-path" data-path="${escapeHtml(s.folderPath)}" title="${t('title_open_folder')}">${ICON.folder}</button>
                                        <button data-act="edit-series" data-brand="${escapeHtml(brandKey)}" data-idx="${idx}" title="${t('title_edit_series')}">${ICON.edit}</button>
                                        <button class="danger" data-act="del-series" data-brand="${escapeHtml(brandKey)}" data-idx="${idx}" title="${t('title_delete_series')}">${ICON.trash}</button>
                                    </div>
                                </div>
                            `).join('')}
                        </div>
                    `}
                </div>
            `;
        }).join('')}
    `;

    if (!detail._wired) {
        detail._wired = true;
        detail.addEventListener('click', onDriverDetailClick);
    }
}

async function onDriverDetailClick(e) {
    const btn = e.target.closest('button[data-act]');
    if (!btn) return;
    const act = btn.dataset.act;
    const brandKey = btn.dataset.brand;
    const idx = btn.dataset.idx != null ? parseInt(btn.dataset.idx, 10) : -1;
    const typeKey = selectedDriverType;

    switch (act) {
        case 'add-series': openDriverAddSeries(typeKey, brandKey); break;
        case 'add-brand': openDriverAddBrand(typeKey); break;
        case 'edit-type': openDriverEditType(typeKey); break;
        case 'edit-brand': openDriverEditBrand(typeKey, brandKey); break;
        case 'edit-series': openDriverEditSeries(typeKey, brandKey, idx); break;
        case 'del-type': await deleteDriverType(typeKey); break;
        case 'del-brand': await deleteDriverBrand(typeKey, brandKey); break;
        case 'del-series': await deleteDriverSeries(typeKey, brandKey, idx); break;
        case 'copy-path': {
            const rawPath = (btn.dataset.path || '').trim();
            if (!rawPath) { showToast(t('err_no_copy'), true); break; }
            // Backend rozwiązuje ścieżkę względną z JSON-a do absolutnej Windows-owej
            let resolved = rawPath;
            try {
                if (window.api && typeof window.api.resolve_path === 'function') {
                    resolved = await window.api.resolve_path(rawPath);
                }
            } catch (e) {
                // fallback: kopiujemy surową ścieżkę
            }
            const ok = await copyToClipboard(resolved || rawPath);
            showToast(ok ? t('toast_path_copied') : t('err_copy_failed'), !ok);
            break;
        }
        case 'open-path': if (window.openFolder) window.openFolder(btn.dataset.path); break;
    }
}

// --- Driver editor (modal) ---

function showDriverFields(mode) {
    driverEditorMode = mode;
    document.getElementById('driver-fields-series').style.display = mode === 'series' ? 'block' : 'none';
    document.getElementById('driver-fields-type').style.display = mode === 'type' ? 'block' : 'none';
    document.getElementById('driver-fields-brand').style.display = mode === 'brand' ? 'block' : 'none';
}

function refreshDriverSelects(seriesTypeSel, seriesBrandSel, brandTypeSel) {
    const typeOpts = sortedTypeEntries()
        .map(([k, t]) => `<option value="${escapeHtml(k)}">${escapeHtml(t)}</option>`)
        .join('');
    if (seriesTypeSel) seriesTypeSel.innerHTML = '<option value="">— wybierz —</option>' + typeOpts;
    if (brandTypeSel) brandTypeSel.innerHTML = '<option value="">— wybierz —</option>' + typeOpts;

    function fillBrandsFor(typeKey, brandSel, currentBrandKey) {
        brandSel.innerHTML = '<option value="">— wybierz —</option>';
        if (!typeKey || !driversData.types[typeKey]) return;
        const opts = sortedBrandEntries(typeKey)
            .map(([k, t]) => {
                const sel = currentBrandKey && k === currentBrandKey ? ' selected' : '';
                return `<option value="${escapeHtml(k)}"${sel}>${escapeHtml(t)}</option>`;
            })
            .join('');
        brandSel.innerHTML += opts;
    }
    return fillBrandsFor;
}

function openDriverEditor(title, subtitle) {
    document.getElementById('driver-editor-title').textContent = title;
    document.getElementById('driver-editor-subtitle').textContent = subtitle || '';
    document.getElementById('driver-editor-overlay').classList.add('active');
}

function closeDriverEditor() {
    document.getElementById('driver-editor-overlay').classList.remove('active');
    driverEditContext = null;
}

function openDriverAddSeries(presetType, presetBrand) {
    driverEditContext = null;
    showDriverFields('series');
    document.getElementById('drv-series-name').value = '';
    document.getElementById('drv-series-folder').value = '';
    const typeSel = document.getElementById('drv-series-type');
    const brandSel = document.getElementById('drv-series-brand');
    const fillBrandsFor = refreshDriverSelects(typeSel, brandSel, null);
    typeSel.value = presetType || selectedDriverType || '';
    fillBrandsFor(typeSel.value, brandSel, presetBrand || '');
    typeSel.onchange = () => fillBrandsFor(typeSel.value, brandSel, '');
    openDriverEditor(t('drv_editor_add_series'), t('drv_editor_series_hint'));
}

function openDriverEditSeries(typeKey, brandKey, seriesIndex) {
    const series = driversData.types[typeKey]?.brands?.[brandKey]?.series?.[seriesIndex];
    if (!series) return;
    driverEditContext = { typeKey, brandKey, seriesIndex };
    showDriverFields('series');
    document.getElementById('drv-series-name').value = series.name || '';
    document.getElementById('drv-series-folder').value = series.folderPath || '';
    const typeSel = document.getElementById('drv-series-type');
    const brandSel = document.getElementById('drv-series-brand');
    const fillBrandsFor = refreshDriverSelects(typeSel, brandSel, brandKey);
    typeSel.value = typeKey;
    fillBrandsFor(typeKey, brandSel, brandKey);
    typeSel.onchange = () => fillBrandsFor(typeSel.value, brandSel, '');
    openDriverEditor(t('drv_editor_edit_series'), t('drv_editor_edit_series_hint'));
}

function openDriverAddType() {
    driverEditContext = null;
    showDriverFields('type');
    document.getElementById('drv-type-key').value = '';
    document.getElementById('drv-type-title').value = '';
    openDriverEditor(t('drv_editor_add_type'), t('drv_editor_add_type_hint'));
}

function openDriverEditType(typeKey) {
    // Nie `t` — przysłoniłoby globalną funkcję tłumaczeń t() użytą niżej.
    const typeData = driversData.types[typeKey];
    if (!typeData) return;
    driverEditContext = { typeKey };
    showDriverFields('type');
    document.getElementById('drv-type-key').value = typeKey;
    document.getElementById('drv-type-title').value = typeData.title || '';
    openDriverEditor(t('drv_editor_edit_type'), t('drv_editor_edit_type_hint'));
}

function openDriverAddBrand(presetType) {
    driverEditContext = null;
    showDriverFields('brand');
    document.getElementById('drv-brand-key').value = '';
    document.getElementById('drv-brand-title').value = '';
    const typeSel = document.getElementById('drv-brand-type');
    const fillBrandsFor = refreshDriverSelects(null, null, typeSel);
    typeSel.value = presetType || selectedDriverType || '';
    openDriverEditor(t('drv_editor_add_brand'), t('drv_editor_add_brand_hint'));
}

function openDriverEditBrand(typeKey, brandKey) {
    const b = driversData.types[typeKey]?.brands?.[brandKey];
    if (!b) return;
    driverEditContext = { typeKey, brandKey };
    showDriverFields('brand');
    document.getElementById('drv-brand-key').value = brandKey;
    document.getElementById('drv-brand-title').value = b.title || '';
    const typeSel = document.getElementById('drv-brand-type');
    const fillBrandsFor = refreshDriverSelects(null, null, typeSel);
    typeSel.value = typeKey;
    openDriverEditor(t('drv_editor_edit_brand'), t('drv_editor_edit_brand_hint'));
}

function slugify(s) {
    return String(s || '')
        .toLowerCase()
        .replace(/ą/g, 'a').replace(/ć/g, 'c').replace(/ę/g, 'e')
        .replace(/ł/g, 'l').replace(/ń/g, 'n').replace(/ó/g, 'o')
        .replace(/ś/g, 's').replace(/ź/g, 'z').replace(/ż/g, 'z')
        .replace(/[^a-z0-9]+/g, '_')
        .replace(/^_+|_+$/g, '');
}

async function saveDriverEditor() {
    if (driverEditorMode === 'series') {
        const typeKey = document.getElementById('drv-series-type').value.trim();
        const brandKey = document.getElementById('drv-series-brand').value.trim();
        const name = document.getElementById('drv-series-name').value.trim();
        const folderPath = document.getElementById('drv-series-folder').value.trim();
        if (!typeKey || !brandKey) { alert('Wybierz typ i producenta.'); return; }
        if (!name) { alert(t('val_series_name')); return; }
        if (!folderPath) { alert(t('val_folder_path')); return; }
        if (!driversData.types[typeKey]) { alert(t('err_type_not_exists')); return; }
        if (!driversData.types[typeKey].brands[brandKey]) { alert(t('err_brand_not_exists')); return; }

        if (driverEditContext) {
            const ctx = driverEditContext;
            const oldBrand = driversData.types[ctx.typeKey].brands[ctx.brandKey];
            const [moved] = oldBrand.series.splice(ctx.seriesIndex, 1);
            moved.name = name;
            moved.folderPath = folderPath;
            driversData.types[typeKey].brands[brandKey].series.push(moved);
        } else {
            driversData.types[typeKey].brands[brandKey].series.push({ name, folderPath });
        }
    } else if (driverEditorMode === 'type') {
        let key = document.getElementById('drv-type-key').value.trim();
        const title = document.getElementById('drv-type-title').value.trim();
        if (!title) { alert(t('val_display_name')); return; }
        if (!key) key = slugify(title);
        if (!/^[a-z0-9_]+$/.test(key)) {
            alert(t('val_type_id_format'));
            return;
        }
        if (driverEditContext) {
            const oldKey = driverEditContext.typeKey;
            if (oldKey !== key && driversData.types[key]) {
                alert('Typ o ID "' + key + '" ' + (typeof t === 'function' ? t('err_already_exists') : 'już istnieje.')); return;
            }
            const oldVal = driversData.types[oldKey];
            delete driversData.types[oldKey];
            driversData.types[key] = { title, brands: oldVal.brands || {} };
            driversData.types[key].title = title;
            if (selectedDriverType === oldKey) selectedDriverType = key;
        } else {
            if (driversData.types[key]) { alert('Typ o ID "' + key + '" ' + (typeof t === 'function' ? t('err_already_exists') : 'już istnieje.')); return; }
            driversData.types[key] = { title, brands: {} };
            selectedDriverType = key;
        }
    } else if (driverEditorMode === 'brand') {
        const typeKey = document.getElementById('drv-brand-type').value.trim();
        let key = document.getElementById('drv-brand-key').value.trim();
        const title = document.getElementById('drv-brand-title').value.trim();
        if (!typeKey) { alert(typeof t === 'function' ? t('val_select_type') : 'Wybierz typ.'); return; }
        if (!title) { alert(t('val_name_required')); return; }
        if (!driversData.types[typeKey]) { alert(t('err_type_not_exists')); return; }
        if (!driversData.types[typeKey].brands) driversData.types[typeKey].brands = {};

        if (driverEditContext) {
            const ctx = driverEditContext;
            const oldBrands = driversData.types[ctx.typeKey].brands;
            const sameType = ctx.typeKey === typeKey;
            if (!sameType && driversData.types[typeKey].brands[key]) {
                alert('Producent o ID "' + key + '" ' + (typeof t === 'function' ? t('err_brand_exists_in_type') : 'już istnieje w tym typie.')); return;
            }
            if (sameType && ctx.brandKey !== key && driversData.types[typeKey].brands[key]) {
                alert('Producent o ID "' + key + '" ' + (typeof t === 'function' ? t('err_brand_exists_in_type') : 'już istnieje w tym typie.')); return;
            }
            const oldVal = oldBrands[ctx.brandKey];
            if (ctx.typeKey !== typeKey) delete oldBrands[ctx.brandKey];
            else if (ctx.brandKey !== key) delete oldBrands[ctx.brandKey];
            driversData.types[typeKey].brands[key] = { title, series: oldVal.series || [] };
        } else {
            if (driversData.types[typeKey].brands[key]) {
                alert('Producent o ID "' + key + '" ' + (typeof t === 'function' ? t('err_brand_exists_in_type') : 'już istnieje w tym typie.')); return;
            }
            driversData.types[typeKey].brands[key] = { title, series: [] };
        }
    }

    const ok = await saveDrivers();
    if (ok) { closeDriverEditor(); renderDriversPane(); }
}

async function deleteDriverType(typeKey) {
    const t = driversData.types[typeKey];
    if (!t) return;
    const seriesCount = Object.values(t.brands || {}).reduce((sum, b) => sum + (b.series || []).length, 0);
    const msg = `Usunąć typ „${t.title}” wraz z ${Object.keys(t.brands || {}).length} producentami i ${seriesCount} seriami?`;
    if (!confirm(msg)) return;
    delete driversData.types[typeKey];
    if (selectedDriverType === typeKey) selectedDriverType = null;
    if (await saveDrivers()) renderDriversPane();
}

async function deleteDriverBrand(typeKey, brandKey) {
    const b = driversData.types[typeKey]?.brands?.[brandKey];
    if (!b) return;
    const seriesCount = (b.series || []).length;
    if (!confirm(`Usunąć producenta „${b.title}” wraz z ${seriesCount} seriami?`)) return;
    delete driversData.types[typeKey].brands[brandKey];
    if (await saveDrivers()) renderDriversPane();
}

async function deleteDriverSeries(typeKey, brandKey, seriesIndex) {
    const list = driversData.types[typeKey]?.brands?.[brandKey]?.series;
    if (!list || !list[seriesIndex]) return;
    const name = list[seriesIndex].name;
    if (!confirm(`Usunąć serię „${name}”?`)) return;
    list.splice(seriesIndex, 1);
    if (await saveDrivers()) renderDriversPane();
}
