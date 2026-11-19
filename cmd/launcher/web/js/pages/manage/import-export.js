// ============================================
// IMPORT/EXPORT
// ============================================

// Default import/export folders come from the backend, derived from the drive
// CoreDeck.exe runs from right now (the pendrive may be L:, J:, ...). Never
// hardcode a drive letter here.
let impexPaths = null;

async function loadImpexPaths() {
    if (impexPaths) return impexPaths;
    try {
        if (window.api && typeof window.api.get_portable_paths === 'function') {
            const p = await window.api.get_portable_paths();
            if (p && !p.error && p.root) impexPaths = p;
        }
    } catch (e) {
        console.warn('get_portable_paths failed:', e);
    }
    return impexPaths;
}

function impexDefault(key) {
    return (impexPaths && impexPaths[key]) || '';
}

function impexJoin(dir, name) {
    return dir ? dir.replace(/[\\/]+$/, '') + '\\' + name : '';
}

// Fill still-empty inputs once the drive-derived defaults are known.
async function fillImpexDefaults(fields) {
    await loadImpexPaths();
    for (const [id, attr, valueFn] of fields) {
        const el = document.getElementById(id);
        if (el && !el[attr]) el[attr] = valueFn();
    }
}

function renderImportExportPane() {
    if (activeManageTab !== 'import-export') return;
    loadImpexPaths();

    // Initialize tabs
    document.querySelectorAll('.impex-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const target = tab.dataset.impexTab;
            document.querySelectorAll('.impex-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            document.querySelectorAll('.impex-pane').forEach(p => p.classList.remove('active'));
            document.getElementById('impex-' + target + '-pane').classList.add('active');

            if (target === 'export') renderExportPane();
            else if (target === 'import') renderImportPane();
            else if (target === 'backups') renderBackupsPane();
        });
    });

    // Render initial export pane
    renderExportPane();
}

function renderExportPane() {
    const pane = document.getElementById('impex-export-pane');
    if (!pane) return;

    pane.innerHTML = `
        <div class="impex-section">
            <div class="impex-section-title">${t('export_title')}</div>
            <div class="impex-section-body">
                <div class="impex-checkbox-group">
                    <label class="impex-checkbox-label">
                        <input type="checkbox" id="export-programs-check" checked>
                        <span>${t('export_programs')}</span>
                    </label>
                    <label class="impex-checkbox-label">
                        <input type="checkbox" id="export-drivers-check" checked>
                        <span>${t('export_drivers')}</span>
                    </label>
                </div>
            </div>
        </div>

        <div class="impex-section" id="export-programs-section">
            <div class="impex-section-title">${tn('count_mng_export_programs_available', allPrograms.length)}</div>
            <div class="impex-section-body">
                <div class="impex-field">
                    <label class="impex-label">${t('label_search')}</label>
                    <input type="text" class="impex-input" id="export-programs-search" placeholder="${escapeHtml(t('ph_filter_programs'))}">
                </div>
                <div class="impex-checkbox-group">
                    <label class="impex-checkbox-label">
                        <input type="checkbox" id="export-programs-select-all">
                        <span style="font-weight:600;">${t('btn_select_all')}</span>
                    </label>
                </div>
                <div class="impex-checkbox-group" id="export-programs-list"></div>
            </div>
        </div>

        <div class="impex-section" id="export-drivers-section">
            <div class="impex-section-title">${t('export_drivers')}</div>
            <div class="impex-section-body">
                <div class="impex-checkbox-group">
                    <label class="impex-checkbox-label">
                        <input type="checkbox" id="export-drivers-select-all">
                        <span style="font-weight:600;">${t('btn_select_all_types')}</span>
                    </label>
                </div>
                <div class="impex-checkbox-group" id="export-drivers-list"></div>
            </div>
        </div>

        <div class="impex-section">
            <div class="impex-section-title">${t('export_options')}</div>
            <div class="impex-section-body">
                <div class="impex-field">
                    <label class="impex-label">${t('label_format')}</label>
                    <div style="display:flex;gap:16px;">
                        <label class="impex-checkbox-label" style="cursor:pointer;">
                            <input type="radio" name="export-format" value="folder" checked>
                            <span>${t('format_folder')}</span>
                        </label>
                        <label class="impex-checkbox-label" style="cursor:pointer;">
                            <input type="radio" name="export-format" value="zip">
                            <span>ZIP</span>
                        </label>
                    </div>
                </div>
                <div class="impex-checkbox-group">
                    <label class="impex-checkbox-label">
                        <input type="checkbox" id="export-include-files" checked>
                        <span>${t('export_option_files')}</span>
                    </label>
                </div>
                <div class="impex-field">
                    <label class="impex-label">${t('label_target_folder')}</label>
                    <div style="display:flex;gap:8px;">
                        <input type="text" class="impex-input" id="export-target" value="" style="flex:1;">
                        <button class="btn btn-secondary" onclick="selectExportTarget()">${t('btn_browse')}</button>
                    </div>
                </div>
            </div>
        </div>

        <div class="impex-section">
            <div class="impex-section-title">${t('summary_title')}</div>
            <div class="impex-summary" id="export-summary">
                <div class="impex-summary-line"><span>${t('summary_programs')}</span><span id="export-summary-programs">${tn('count_mng_selected', 0)}</span></div>
                <div class="impex-summary-line"><span>${t('summary_drivers')}</span><span id="export-summary-drivers">${tn('count_mng_selected', 0)}</span></div>
                <div class="impex-summary-line"><span>${t('summary_size')}</span><span id="export-summary-size">—</span></div>
                <div id="export-summary-warnings"></div>
            </div>
        </div>

        <div class="impex-actions">
            <button class="btn btn-primary" onclick="startExport()">${t('btn_export')}</button>
        </div>
    `;

    renderExportProgramsList();
    renderExportDriversList();
    updateExportSummary();
    fillImpexDefaults([
        ['export-target', 'value', () => impexJoin(impexDefault('export'), 'Pack_' + new Date().toISOString().split('T')[0])],
    ]);

    // Wire up events
    document.getElementById('export-programs-check')?.addEventListener('change', toggleExportProgramsSection);
    document.getElementById('export-drivers-check')?.addEventListener('change', toggleExportDriversSection);
    document.getElementById('export-programs-search')?.addEventListener('input', renderExportProgramsList);
    document.getElementById('export-programs-select-all')?.addEventListener('change', toggleAllExportPrograms);
    document.getElementById('export-drivers-select-all')?.addEventListener('change', toggleAllExportDrivers);
}

function renderExportProgramsList() {
    const list = document.getElementById('export-programs-list');
    if (!list) return;

    const search = (document.getElementById('export-programs-search')?.value || '').toLowerCase();
    const filtered = allPrograms.filter(p => {
        if (!search) return true;
        const hay = [p.name, p.description, ...(p.tags || [])].join(' ').toLowerCase();
        return hay.includes(search);
    });

    list.innerHTML = filtered.map(p => {
        const isPage = p.type === 'link';
        const typeLabel = isPage ? escapeHtml(t('chip_page_label')) : escapeHtml(p.installType || 'other');
        const tagClass = badgeClass(isPage ? 'page' : (p.installType || 'other'));
        return `
        <label class="impex-checkbox-label impex-program-row">
            <input type="checkbox" class="export-program-cb" data-name="${escapeHtml(p.name)}">
            <span class="impex-program-icon">${renderProgramIcon(p)}</span>
            <span class="impex-program-info">
                <span class="impex-program-name">${escapeHtml(p.name)}</span>
                ${p.description ? `<span class="impex-program-desc">${escapeHtml(p.description)}</span>` : ''}
            </span>
            <span class="${tagClass}">${typeLabel}</span>
        </label>`;
    }).join('');

    list.querySelectorAll('.export-program-cb').forEach(cb => {
        cb.addEventListener('change', updateExportSummary);
    });
}

function renderExportDriversList() {
    const list = document.getElementById('export-drivers-list');
    if (!list) return;

    const types = Object.entries(driversData.types || {});
    list.innerHTML = types.map(([typeKey, type]) => {
        const brands = Object.entries(type.brands || {});
        const brandHtml = brands.map(([brandKey, brand]) => {
            const series = brand.series || [];
            const seriesHtml = series.map(s => `
                <label class="impex-checkbox-label impex-series-row" style="padding-left:40px;">
                    <input type="checkbox" class="export-driver-series-cb" data-type="${escapeHtml(typeKey)}" data-brand="${escapeHtml(brandKey)}" data-series="${escapeHtml(s.name)}">
                    <span style="font-size:var(--fs-xs);">${escapeHtml(s.name)}</span>
                </label>
            `).join('');
            return `
                <label class="impex-checkbox-label impex-brand-row" style="padding-left:24px;">
                    <input type="checkbox" class="export-driver-brand-cb" data-type="${escapeHtml(typeKey)}" data-brand="${escapeHtml(brandKey)}">
                    <span style="font-weight:500;">${escapeHtml(brand.title || brandKey)} <span style="color:var(--text-muted);font-size:var(--fs-2xs);">(${series.length})</span></span>
                </label>
                ${seriesHtml}
            `;
        }).join('');
        return `
            <label class="impex-checkbox-label impex-type-row">
                <input type="checkbox" class="export-driver-type-cb" data-type="${escapeHtml(typeKey)}">
                <span style="font-weight:600;">${escapeHtml(type.title || typeKey)}</span>
            </label>
            ${brandHtml}
        `;
    }).join('');

    // Wire up checkbox logic: type → brands → series cascading
    list.querySelectorAll('.export-driver-type-cb').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const typeKey = cb.dataset.type;
            const checked = cb.checked;
            list.querySelectorAll(`.export-driver-brand-cb[data-type="${typeKey}"]`).forEach(bcb => {
                bcb.checked = checked;
                bcb.dispatchEvent(new Event('change'));
            });
            updateExportSummary();
        });
    });

    list.querySelectorAll('.export-driver-brand-cb').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const typeKey = cb.dataset.type;
            const brandKey = cb.dataset.brand;
            const checked = cb.checked;
            list.querySelectorAll(`.export-driver-series-cb[data-type="${typeKey}"][data-brand="${brandKey}"]`).forEach(scb => {
                scb.checked = checked;
            });
            // Sprawdź czy wszystkie brand-e w type są zaznaczone
            const allBrands = list.querySelectorAll(`.export-driver-brand-cb[data-type="${typeKey}"]`);
            const allChecked = Array.from(allBrands).every(b => b.checked);
            const typeCb = list.querySelector(`.export-driver-type-cb[data-type="${typeKey}"]`);
            if (typeCb) typeCb.checked = allChecked;
            updateExportSummary();
        });
    });

    list.querySelectorAll('.export-driver-series-cb').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const typeKey = cb.dataset.type;
            const brandKey = cb.dataset.brand;
            // Sprawdź czy wszystkie serie w brand są zaznaczone
            const allSeries = list.querySelectorAll(`.export-driver-series-cb[data-type="${typeKey}"][data-brand="${brandKey}"]`);
            const allChecked = Array.from(allSeries).every(s => s.checked);
            const brandCb = list.querySelector(`.export-driver-brand-cb[data-type="${typeKey}"][data-brand="${brandKey}"]`);
            if (brandCb) brandCb.checked = allChecked;
            // Sprawdź czy type
            const allBrands = list.querySelectorAll(`.export-driver-brand-cb[data-type="${typeKey}"]`);
            const allBrandsChecked = Array.from(allBrands).every(b => b.checked);
            const typeCb = list.querySelector(`.export-driver-type-cb[data-type="${typeKey}"]`);
            if (typeCb) typeCb.checked = allBrandsChecked;
            updateExportSummary();
        });
    });
}

function toggleExportProgramsSection(e) {
    const section = document.getElementById('export-programs-section');
    if (section) section.style.display = e.target.checked ? 'block' : 'none';
    updateExportSummary();
}

function toggleExportDriversSection(e) {
    const section = document.getElementById('export-drivers-section');
    if (section) section.style.display = e.target.checked ? 'block' : 'none';
    updateExportSummary();
}

function toggleAllExportPrograms(e) {
    document.querySelectorAll('.export-program-cb').forEach(cb => cb.checked = e.target.checked);
    updateExportSummary();
}

function toggleAllExportDrivers(e) {
    const checked = e.target.checked;
    document.querySelectorAll('.export-driver-type-cb').forEach(cb => cb.checked = checked);
    document.querySelectorAll('.export-driver-brand-cb').forEach(cb => cb.checked = checked);
    document.querySelectorAll('.export-driver-series-cb').forEach(cb => cb.checked = checked);
    updateExportSummary();
}

function updateExportSummary() {
    const progCount = document.querySelectorAll('.export-program-cb:checked').length;
    // Zlicz sterowniki: series checkboxes są najniższym poziomem
    const seriesChecked = document.querySelectorAll('.export-driver-series-cb:checked').length;
    const brandChecked = document.querySelectorAll('.export-driver-brand-cb:checked').length;
    const typeChecked = document.querySelectorAll('.export-driver-type-cb:checked').length;
    // Jeśli jest granular — liczymy serie; jeśli stare — typy
    const drvCount = seriesChecked > 0 ? seriesChecked : (brandChecked > 0 ? brandChecked : typeChecked);
    setText('export-summary-programs', tn('count_mng_selected', progCount));
    setText('export-summary-drivers', tn('count_mng_selected', drvCount));

    // Debounced walidacja — sprawdź foldery i szacuj rozmiar
    clearTimeout(updateExportSummary._timer);
    updateExportSummary._timer = setTimeout(() => runExportValidation(), 400);
}

async function runExportValidation() {
    const summaryEl = document.getElementById('export-summary');
    if (!summaryEl) return;
    const sizeEl = document.getElementById('export-summary-size');
    const warnEl = document.getElementById('export-summary-warnings');
    const progCount = document.querySelectorAll('.export-program-cb:checked').length;
    // `.export-driver-cb` nie istnieje od czasu rozbicia eksportu sterowników na
    // trzy poziomy (type/brand/series) — ten selektor nic nie łapał, więc
    // drvCount był zawsze 0 i walidacja "tylko sterowniki" nigdy nie startowała.
    // Liczymy typy, bo to one trafiają do request.driverTypes niżej.
    const drvCount = document.querySelectorAll('.export-driver-type-cb:checked').length;
    if (progCount === 0 && drvCount === 0) {
        if (sizeEl) sizeEl.textContent = '—';
        if (warnEl) warnEl.innerHTML = '';
        return;
    }
    if (sizeEl) sizeEl.textContent = t('calc_loading');

    const includePrograms = document.getElementById('export-programs-check')?.checked !== false;
    const includeDrivers = document.getElementById('export-drivers-check')?.checked !== false;

    const driverSelections = {};
    if (includeDrivers) {
        document.querySelectorAll('.export-driver-type-cb:checked').forEach(typeCb => {
            const typeKey = typeCb.dataset.type;
            const brands = {};
            document.querySelectorAll(`.export-driver-brand-cb[data-type="${typeKey}"]:checked`).forEach(brandCb => {
                const brandKey = brandCb.dataset.brand;
                const seriesNames = [];
                document.querySelectorAll(`.export-driver-series-cb[data-type="${typeKey}"][data-brand="${brandKey}"]:checked`).forEach(scb => {
                    seriesNames.push(scb.dataset.series);
                });
                brands[brandKey] = { series: seriesNames };
            });
            driverSelections[typeKey] = { brands };
        });
    }

    const request = {
        programNames: includePrograms ? Array.from(document.querySelectorAll('.export-program-cb:checked')).map(cb => cb.dataset.name) : [],
        driverTypes: includeDrivers ? Array.from(document.querySelectorAll('.export-driver-type-cb:checked')).map(cb => cb.dataset.type) : [],
        driverSelections,
        includeFiles: document.getElementById('export-include-files')?.checked || false,
    };
    try {
        const result = await window.api.validate_export(JSON.stringify(request));
        if (sizeEl) sizeEl.textContent = result.estimatedSize ? humanSizeBytes(result.estimatedSize) : '—';
        if (warnEl) {
            if (result.warnings && result.warnings.length > 0) {
                warnEl.innerHTML = result.warnings.map(w =>
                    `<div class="impex-warning">⚠ ${t('mng_export_warn_missing_folder', { item: escapeHtml(w.item), path: '<code>' + escapeHtml(w.path) + '</code>' })}</div>`
                ).join('');
            } else {
                warnEl.innerHTML = '';
            }
        }
    } catch (e) {
        if (sizeEl) sizeEl.textContent = '—';
    }
}

// Log/error line of an export or import (impex.Line: message + code + params).
function impexLineText(line) {
    if (typeof line === 'string') return line;
    return line ? tMsg(line.code, line.params, line.message) : '';
}

function humanSizeBytes(b) {
    if (b >= 1073741824) return (b / 1073741824).toFixed(1) + ' GB';
    if (b >= 1048576) return (b / 1048576).toFixed(1) + ' MB';
    if (b >= 1024) return Math.round(b / 1024) + ' KB';
    return b + ' B';
}

function selectExportTarget() {
    (async () => {
        const format = document.querySelector('input[name="export-format"]:checked')?.value || 'folder';
        const input = document.getElementById('export-target');
        const current = input?.value || impexDefault('export');
        let picked = '';
        if (format === 'zip') {
            picked = await window.api.pick_save_file(
                t('save_zip_dialog'),
                getParentDir(current),
                t('file_filter_zip'),
                current.toLowerCase().endsWith('.zip') ? getBaseName(current) : getBaseName(current) + '.zip'
            );
        } else {
            picked = await window.api.pick_folder(t('pick_export_folder'), current);
        }
        if (picked && input) input.value = picked;
    })();
}

async function startExport() {
    const includePrograms = document.getElementById('export-programs-check')?.checked !== false;
    const includeDrivers = document.getElementById('export-drivers-check')?.checked !== false;
    const selectedPrograms = includePrograms
        ? Array.from(document.querySelectorAll('.export-program-cb:checked')).map(cb => cb.dataset.name)
        : [];
    const selectedDrivers = includeDrivers
        ? Array.from(document.querySelectorAll('.export-driver-type-cb:checked')).map(cb => cb.dataset.type)
        : [];

    if (selectedPrograms.length === 0 && selectedDrivers.length === 0) {
        showToast(t('export_err_no_selection'), true);
        return;
    }

    const includeFiles = document.getElementById('export-include-files')?.checked || false;
    const format = document.querySelector('input[name="export-format"]:checked')?.value || 'folder';
    const targetPath = document.getElementById('export-target')?.value || '';

    if (!targetPath) {
        showToast(t('export_err_no_target'), true);
        return;
    }

    // Zbierz granularne zaznaczenie sterowników
    const driverSelections = {};
    if (includeDrivers) {
        const typeCbs = document.querySelectorAll('.export-driver-type-cb:checked');
        typeCbs.forEach(typeCb => {
            const typeKey = typeCb.dataset.type;
            const brands = {};
            document.querySelectorAll(`.export-driver-brand-cb[data-type="${typeKey}"]:checked`).forEach(brandCb => {
                const brandKey = brandCb.dataset.brand;
                const seriesNames = [];
                document.querySelectorAll(`.export-driver-series-cb[data-type="${typeKey}"][data-brand="${brandKey}"]:checked`).forEach(scb => {
                    seriesNames.push(scb.dataset.series);
                });
                // Jeśli brand jest zaznaczony ale bez zaznaczonych serii → wszystkie serie
                brands[brandKey] = { series: seriesNames };
            });
            driverSelections[typeKey] = { brands };
        });
    }

    const request = {
        programNames: selectedPrograms,
        driverTypes: selectedDrivers,
        driverSelections,
        includeFiles,
        format,
        targetPath
    };

    // Pokaż progress overlay
    showProgress(t('progress_exporting'), [
        t('progress_preparing'),
        includeFiles ? t('progress_copying') : t('progress_saving_meta'),
        format === 'zip' ? t('progress_packing_zip') : '',
    ].filter(Boolean));

    try {
        const result = await window.api.export_pack(JSON.stringify(request));

        hideProgress();

        if (result.success) {
            showToast(t('mng_export_done', {
                programs: tn('count_programs', result.programCount || 0),
                drivers: tn('count_mng_drivers', result.driverCount || 0),
                path: result.exportedPath || '',
            }), false);
            // Pokaż log eksportu
            if (result.log && result.log.length > 0) {
                showExportLog(result);
            }
            // Odśwież panel backupów, jeśli jest widoczny — nowy eksport trafi na listę.
            const backupsPane = document.getElementById('impex-backups-pane');
            if (backupsPane && backupsPane.classList.contains('active')) {
                renderBackupsPane();
            }
        } else {
            showToast(t('mng_export_failed', { error: tErr(result) }), true);
        }

        if (result.errors && result.errors.length > 0) {
            console.error('Export errors:', result.errors);
            result.errors.forEach(e => showToast(impexLineText(e), true));
        }
    } catch (e) {
        hideProgress();
        showToast(t('mng_export_failed', { error: String(e) }), true);
    }
}

function showExportLog(result) {
    const summary = document.getElementById('export-summary');
    if (!summary) return;
    // Usuń stary log jeśli jest
    const old = document.getElementById('export-log-box');
    if (old) old.remove();
    const logHtml = `
        <div id="export-log-box" class="impex-log-box">
            <div class="impex-log-title">${t('export_log')}</div>
            ${(result.log || []).map(line => `<div class="impex-log-line">✓ ${escapeHtml(impexLineText(line))}</div>`).join('')}
            ${result.exportedPath ? `<div class="impex-log-line" style="color:var(--accent-primary);">→ ${escapeHtml(result.exportedPath)}</div>` : ''}
        </div>
    `;
    summary.insertAdjacentHTML('afterend', logHtml);
}

function renderImportPane() {
    const pane = document.getElementById('impex-import-pane');
    if (!pane) return;

    pane.innerHTML = `
        <div class="impex-section">
            <div class="impex-section-title">${t('import_title')}</div>
            <div class="impex-section-body">
                <div class="impex-field">
                    <label class="impex-label">${t('label_pack')}</label>
                    <div style="display:flex;gap:8px;">
                        <input type="text" class="impex-input" id="import-pack-path" placeholder="" style="flex:1;">
                        <button class="btn btn-secondary" onclick="selectImportZip()">ZIP</button>
                        <button class="btn btn-secondary" onclick="selectImportFolder()">${t('format_folder')}</button>
                        <button class="btn btn-secondary" onclick="previewImportPack()">${t('btn_check')}</button>
                    </div>
                </div>
            </div>
        </div>

        <div class="impex-section" id="import-content-section" style="display:none;">
            <div class="impex-section-title">${t('import_content')}</div>
            <div class="impex-section-body" id="import-content"></div>
        </div>

        <div class="impex-section">
            <div class="impex-section-title">${t('import_strategy')}</div>
            <div class="impex-section-body">
                <div style="display:flex;flex-direction:column;gap:8px;">
                    <label class="impex-checkbox-label" style="cursor:pointer;">
                        <input type="radio" name="import-strategy" value="skip">
                        <span>${t('strategy_skip')}</span>
                    </label>
                    <label class="impex-checkbox-label" style="cursor:pointer;">
                        <input type="radio" name="import-strategy" value="overwrite" checked>
                        <span>${t('strategy_overwrite')}</span>
                    </label>
                    <label class="impex-checkbox-label" style="cursor:pointer;">
                        <input type="radio" name="import-strategy" value="duplicate">
                        <span>${escapeHtml(t('strategy_duplicate'))}</span>
                    </label>
                </div>
            </div>
        </div>

        <div class="impex-section">
            <div class="impex-section-title">${t('import_target')}</div>
            <div class="impex-section-body">
                <div class="impex-field">
                    <label class="impex-label">${t('import_target_programs')}</label>
                    <div style="display:flex;gap:8px;">
                        <input type="text" class="impex-input" id="import-programs-target" value="" style="flex:1;">
                        <button class="btn btn-secondary" onclick="selectImportProgramsTarget()">${t('btn_browse')}</button>
                    </div>
                </div>
                <div class="impex-field">
                    <label class="impex-label">${t('import_target_drivers')}</label>
                    <div style="display:flex;gap:8px;">
                        <input type="text" class="impex-input" id="import-drivers-target" value="" style="flex:1;">
                        <button class="btn btn-secondary" onclick="selectImportDriversTarget()">${t('btn_browse')}</button>
                    </div>
                </div>
            </div>
        </div>

        <div class="impex-section">
            <div class="impex-section-title">${t('summary_title')}</div>
            <div class="impex-summary" id="import-summary">
                <div class="impex-summary-line"><span>${t('summary_status')}</span><span>${t('summary_select_pack')}</span></div>
            </div>
        </div>

        <div class="impex-actions">
            <button class="btn btn-primary" onclick="startImport()">${t('btn_import')}</button>
        </div>
    `;
    fillImpexDefaults([
        ['import-pack-path', 'placeholder', () => impexJoin(impexDefault('import'), 'Gaming_Pack.zip')],
        ['import-programs-target', 'value', () => impexDefault('programs')],
        ['import-drivers-target', 'value', () => impexDefault('drivers')],
    ]);
}

// ============================================
// BACKUPS — historia eksportów
// ============================================

function formatBackupDate(iso) {
    if (!iso) return '—';
    const d = new Date(iso);
    if (isNaN(d.getTime())) return iso;
    const lang = typeof getCurrentLang === 'function' ? getCurrentLang() : undefined;
    const opts = { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' };
    try {
        return new Intl.DateTimeFormat((lang && LANGUAGE_TAGS[lang]) || lang, opts).format(d);
    } catch (_) {
        return new Intl.DateTimeFormat(undefined, opts).format(d);
    }
}

async function renderBackupsPane() {
    const pane = document.getElementById('impex-backups-pane');
    if (!pane) return;

    let records = [];
    try {
        records = await window.api.list_export_history();
    } catch (e) {
        records = [];
    }

    const header = `
        <div class="impex-section">
            <div class="bk-header">
                <div class="bk-header-main">
                    <h3>${t('backups_title')}</h3>
                    <p class="bk-header-sub">${tn('count_mng_backups_subtitle', 50)}</p>
                </div>
                ${records.length > 0 ? `<button class="btn btn-danger" onclick="clearBackupHistory()" title="${t('title_clear_history')}">${t('btn_clear_history')}</button>` : ''}
            </div>
            ${records.length === 0 ? `
                <div class="bk-empty">
                    ${ICON.archive}
                    <div class="bk-empty-title">${t('backups_empty_title')}</div>
                    <div>${t('backups_empty_hint')}</div>
                </div>
            ` : `
                <div class="bk-list">
                    ${records.map(r => renderBackupRow(r)).join('')}
                </div>
            `}
        </div>
    `;

    pane.innerHTML = header;

    // Wire up akcji na wpisach (delegacja).
    if (!pane._wired) {
        pane._wired = true;
        pane.addEventListener('click', onBackupRowClick);
    }
}

function renderBackupRow(r) {
    const isZip = (r.format || '').toLowerCase() === 'zip';
    const icon = isZip ? '🗜️' : '📦';
    const formatLabel = isZip ? 'ZIP' : escapeHtml(t('format_folder'));
    const name = escapeHtml(getBaseName(r.targetPath));
    const path = escapeHtml(r.targetPath || '');
    const date = escapeHtml(formatBackupDate(r.createdAt));
    const progText = tn('count_programs', r.programCount || 0);
    const drvText = tn('count_mng_drivers', r.driverCount || 0);
    const filesNote = r.includeFiles === false ? t('backup_json_only') : t('backup_with_files');
    const id = escapeHtml(r.id || '');

    return `
        <div class="bk-row">
            <div class="bk-icon">${icon}</div>
            <div class="bk-main">
                <div class="bk-name" title="${name}">${name}</div>
                <div class="bk-path" title="${path}">${path}</div>
                <div class="bk-meta">
                    <span style="display:inline-flex;align-items:center;gap:3px;">${ICON.clock} ${date}</span>
                    <span class="bk-meta-sep"></span>
                    <span class="${badgeClass(isZip ? 'installer' : 'portable')}">${formatLabel}</span>
                    <span class="bk-meta-sep"></span>
                    <span class="badge badge-neutral">${progText}</span>
                    <span class="bk-meta-sep"></span>
                    <span class="badge badge-neutral">${drvText}</span>
                    <span class="bk-meta-sep"></span>
                    <span>${filesNote}</span>
                </div>
            </div>
            <div class="bk-actions">
                <button class="bk-action" data-bk-act="open" data-bk-path="${path}" title="${t('title_open_folder')}">${ICON.folder}</button>
                <button class="bk-action" data-bk-act="copy" data-bk-path="${path}" title="${t('title_copy_path')}">${ICON.copy}</button>
                <button class="bk-action" data-bk-act="reimport" data-bk-path="${path}" title="${t('title_reimport')}">${ICON.download}</button>
                <button class="bk-action danger" data-bk-act="remove" data-bk-id="${id}" title="${t('title_remove_list')}">${ICON.trash}</button>
            </div>
        </div>
    `;
}

async function onBackupRowClick(e) {
    const btn = e.target.closest('button[data-bk-act]');
    if (!btn) return;
    const act = btn.dataset.bkAct;
    const path = btn.dataset.bkPath || '';
    const id = btn.dataset.bkId || '';

    switch (act) {
        case 'open': {
            if (!path) { showToast(t('err_no_path'), true); break; }
            if (window.openFolder) window.openFolder(path);
            break;
        }
        case 'copy': {
            if (!path) { showToast(t('err_no_copy'), true); break; }
            const ok = await copyToClipboard(path);
            showToast(ok ? t('toast_path_copied') : t('err_copy_failed'), !ok);
            break;
        }
        case 'reimport': {
            if (!path) { showToast(t('err_no_pack_path'), true); break; }
            // Przełącz na tab Import i wstaw ścieżkę do pola paczki.
            const importTab = document.querySelector('.impex-tab[data-impex-tab="import"]');
            if (importTab) importTab.click();
            // renderImportPane przebudowuje DOM — poczekaj na klatkę i wstaw + podgląd.
            requestAnimationFrame(async () => {
                const input = document.getElementById('import-pack-path');
                if (input) {
                    input.value = path;
                    await previewImportPack();
                }
            });
            showToast(t('toast_loaded_import'), false);
            break;
        }
        case 'remove': {
            if (!id) { showToast(t('err_no_entry_id'), true); break; }
            try {
                const res = await window.api.remove_export_history(JSON.stringify({ id }));
                if (res.success) {
                    showToast(t('toast_removed_history'), false);
                    renderBackupsPane();
                } else {
                    showToast(tErr(res) || t('err_cannot_delete'), true);
                }
            } catch (err) {
                showToast(t('mng_err_delete', { error: String(err) }), true);
            }
            break;
        }
    }
}

async function clearBackupHistory() {
    if (!(await uiConfirm(t('confirm_clear_history'), { danger: true, confirmLabel: t('btn_clear') }))) return;
    try {
        const res = await window.api.clear_export_history();
        if (res.success) {
            showToast(t('toast_history_cleared'), false);
            renderBackupsPane();
        } else {
            showToast(tErr(res) || t('err_clearing'), true);
        }
    } catch (err) {
        showToast(t('mng_err_clearing', { error: String(err) }), true);
    }
}

async function selectImportZip() {
    const input = document.getElementById('import-pack-path');
    const picked = await window.api.pick_open_file(
        t('pick_zip'),
        getParentDir(input?.value || impexDefault('export')),
        t('file_filter_zip')
    );
    if (picked && input) {
        input.value = picked;
        await previewImportPack();
    }
}

async function selectImportFolder() {
    const input = document.getElementById('import-pack-path');
    const picked = await window.api.pick_folder(t('pick_pack_folder'), input?.value || impexDefault('export'));
    if (picked && input) {
        input.value = picked;
        await previewImportPack();
    }
}

async function previewImportPack() {
    const packPath = document.getElementById('import-pack-path')?.value || '';
    if (!packPath) {
        showToast(t('import_err_no_pack'), true);
        return;
    }
    try {
        const preview = await window.api.preview_pack(packPath);
        const section = document.getElementById('import-content-section');
        const content = document.getElementById('import-content');
        const summary = document.getElementById('import-summary');
        if (section) section.style.display = 'block';
        if (!preview.success) {
            if (content) content.innerHTML = `<div style="color:var(--danger);">${escapeHtml(tErr(preview) || t('error_no_pack'))}</div>`;
            if (summary) summary.innerHTML = `<div class="impex-summary-line"><span>${t('preview_status')}</span><span>${escapeHtml(tErr(preview) || t('error_generic'))}</span></div>`;
            return;
        }
        if (content) {
            const programs = (preview.programs || []).map(p => `
                <label class="impex-checkbox-label">
                    <input type="checkbox" class="import-program-cb" data-name="${escapeHtml(p.name)}" checked>
                    <span>${escapeHtml(p.name)} <span style="color:var(--text-muted);font-size:var(--fs-xs);">${escapeHtml(p.installType || 'other')}${p.variants ? ', ' + tn('count_mng_variants', p.variants) : ''}</span></span>
                </label>
            `).join('') || `<div style="color:var(--text-muted);">${t('empty_no_programs')}</div>`;
            const drivers = (preview.drivers || []).map(d => `
                <label class="impex-checkbox-label">
                    <input type="checkbox" class="import-driver-cb" data-type="${escapeHtml(d.key)}" checked>
                    <span>${escapeHtml(d.title || d.key)} <span style="color:var(--text-muted);font-size:var(--fs-xs);">${tn('count_mng_brands', d.brands || 0)}, ${tn('count_mng_series', d.series || 0)}</span></span>
                </label>
            `).join('') || `<div style="color:var(--text-muted);">${t('empty_no_driver_types')}</div>`;
            content.innerHTML = `
                <div class="impex-section-title" style="font-size:var(--fs-sm2);margin:0;">${t('mng_heading_programs_count', { count: preview.programCount || 0 })}</div>
                <div class="impex-checkbox-group">${programs}</div>
                <div class="impex-section-title" style="font-size:var(--fs-sm2);margin:12px 0 0;">${t('mng_heading_drivers_count', { count: preview.driverCount || 0 })}</div>
                <div class="impex-checkbox-group">${drivers}</div>
            `;
        }
        if (summary) {
            summary.innerHTML = `
                <div class="impex-summary-line"><span>${t('preview_status')}</span><span>${escapeHtml(tErr(preview))}</span></div>
                <div class="impex-summary-line"><span>${t('summary_programs')}</span><span>${preview.programCount || 0}</span></div>
                <div class="impex-summary-line"><span>${t('summary_drivers')}</span><span>${preview.driverCount || 0}</span></div>
            `;
        }
    } catch (e) {
        showToast(t('mng_err_generic', { error: String(e) }), true);
    }
}

async function selectImportProgramsTarget() {
    const input = document.getElementById('import-programs-target');
    const picked = await window.api.pick_folder(t('pick_programs_target'), input?.value || impexDefault('programs'));
    if (picked && input) input.value = picked;
}

async function selectImportDriversTarget() {
    const input = document.getElementById('import-drivers-target');
    const picked = await window.api.pick_folder(t('pick_drivers_target'), input?.value || impexDefault('drivers'));
    if (picked && input) input.value = picked;
}

async function startImport() {
    const packPath = document.getElementById('import-pack-path')?.value || '';
    const programsTargetDir = document.getElementById('import-programs-target')?.value || '';
    const driversTargetDir = document.getElementById('import-drivers-target')?.value || '';
    const strategy = document.querySelector('input[name="import-strategy"]:checked')?.value || 'overwrite';

    if (!packPath) {
        showToast(t('import_err_no_path'), true);
        return;
    }

    const request = {
        packPath,
        programsTargetDir,
        driversTargetDir,
        strategy,
        importPrograms: true,
        importDrivers: true,
        selectedPrograms: Array.from(document.querySelectorAll('.import-program-cb:checked')).map(cb => cb.dataset.name),
        selectedDrivers: Array.from(document.querySelectorAll('.import-driver-cb:checked')).map(cb => cb.dataset.type)
    };

    // Pokaż progress overlay
    const isZip = packPath.toLowerCase().endsWith('.zip');
    showProgress(t('progress_importing'), [
        isZip ? t('progress_unzipping') : t('mng_progress_reading_pack'),
        t('progress_import_programs'),
        t('progress_import_drivers'),
        t('progress_saving'),
    ]);

    try {
        const result = await window.api.import_pack(JSON.stringify(request));

        hideProgress();

        if (result.success) {
            showToast(t('mng_import_done', {
                programs: tn('count_programs', result.programCount || 0),
                types: tn('count_mng_driver_types', result.driverCount || 0),
            }), false);
            // Pokaż log importu
            showImportLog(result);
            await loadPrograms();
            await loadDrivers();
            renderImportExportPane();
        } else {
            showToast(t('mng_import_failed', { error: tErr(result) }), true);
            if (result.errors) {
                result.errors.forEach(e => showToast(impexLineText(e), true));
                console.error('Import errors:', result.errors);
            }
        }
    } catch (e) {
        hideProgress();
        showToast(t('mng_import_failed', { error: String(e) }), true);
    }
}

function showImportLog(result) {
    const summary = document.getElementById('import-summary');
    if (!summary) return;
    const old = document.getElementById('import-log-box');
    if (old) old.remove();
    const logHtml = `
        <div id="import-log-box" class="impex-log-box">
            <div class="impex-log-title">${t('import_log')}</div>
            ${(result.log || []).map(line => `<div class="impex-log-line">✓ ${escapeHtml(impexLineText(line))}</div>`).join('')}
            ${(result.skipped || []).map(name => `<div class="impex-log-line" style="color:var(--warn);">\u2298 ${escapeHtml(t('mng_log_skipped', { name }))}</div>`).join('')}
            ${(result.overwritten || []).map(name => `<div class="impex-log-line" style="color:var(--info);">\u21bb ${escapeHtml(t('mng_log_overwritten', { name }))}</div>`).join('')}
            ${(result.errors || []).map(e => `<div class="impex-log-line" style="color:var(--danger);">✗ ${escapeHtml(impexLineText(e))}</div>`).join('')}
        </div>
    `;
    summary.insertAdjacentHTML('afterend', logHtml);
}

// ============================================
// PROGRESS OVERLAY
// ============================================

function showProgress(title, steps) {
    let overlay = document.getElementById('progress-overlay');
    if (!overlay) {
        overlay = document.createElement('div');
        overlay.id = 'progress-overlay';
        document.body.appendChild(overlay);
    }
    overlay.innerHTML = `
        <div class="progress-backdrop"></div>
        <div class="progress-modal">
            <div class="progress-spinner"></div>
            <div class="progress-title">${escapeHtml(title)}</div>
            <div class="progress-steps">
                ${(steps || []).map((s, i) => `<div class="progress-step${i === 0 ? ' active' : ''}">${escapeHtml(s)}</div>`).join('')}
            </div>
        </div>
    `;
    overlay.style.display = 'flex';
    // Animacja przechodzenia kroków
    const stepEls = overlay.querySelectorAll('.progress-step');
    let idx = 0;
    if (overlay._stepTimer) clearInterval(overlay._stepTimer);
    if (stepEls.length > 1) {
        overlay._stepTimer = setInterval(() => {
            if (idx < stepEls.length - 1) {
                stepEls[idx].classList.remove('active');
                stepEls[idx].classList.add('done');
                idx++;
                stepEls[idx].classList.add('active');
            }
        }, 1500);
    }
}

function hideProgress() {
    const overlay = document.getElementById('progress-overlay');
    if (overlay) {
        if (overlay._stepTimer) clearInterval(overlay._stepTimer);
        overlay.style.display = 'none';
        overlay.innerHTML = '';
    }
}

function getParentDir(path) {
    if (!path) return '';
    const idx = Math.max(path.lastIndexOf('\\'), path.lastIndexOf('/'));
    return idx > 0 ? path.slice(0, idx) : path;
}

function getBaseName(path) {
    if (!path) return 'Pack_' + new Date().toISOString().split('T')[0];
    const idx = Math.max(path.lastIndexOf('\\'), path.lastIndexOf('/'));
    const name = idx >= 0 ? path.slice(idx + 1) : path;
    return name || ('Pack_' + new Date().toISOString().split('T')[0]);
}
