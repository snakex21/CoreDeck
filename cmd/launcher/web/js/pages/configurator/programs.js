// Map a program installType to a shared badge variant class (style.css).
function installBadgeClass(installType) {
    switch (installType) {
        case 'installer': return 'badge-info';
        case 'portable':  return 'badge-success';
        case 'tool':
        case 'addon':
        case 'collection': return 'badge-purple';
        default: return 'badge-neutral';
    }
}

function renderPrograms() {
    const search = (document.getElementById('search').value || '').toLowerCase();
    const filterType = document.getElementById('filter-type').value;
    const filterCat = document.getElementById('filter-category').value;

    let filtered = allPrograms.filter(p => {
        if (search && !p.name.toLowerCase().includes(search) && !p.description.toLowerCase().includes(search) &&
            !(p.tags || []).some(t => t.toLowerCase().includes(search))) return false;
        if (filterType !== 'all' && p.installType !== filterType) return false;
        if (filterCat !== 'all' && p.category !== filterCat) return false;
        return true;
    });

    const list = document.getElementById('programs-list');
    list.innerHTML = filtered.map((p, i) => {
        const idx = allPrograms.indexOf(p);
        const nameAttr = encodeURIComponent(p.name);
        let versionSelect = '';
        if (p.variants && p.variants.length > 0) {
            if (typeof p.selectedVariant !== 'number' || p.selectedVariant < 0 || p.selectedVariant >= p.variants.length) {
                p.selectedVariant = (typeof p.defaultVariant === 'number' && p.defaultVariant >= 0 && p.defaultVariant < p.variants.length) ? p.defaultVariant : 0;
            }
            const opts = p.variants.map((v, vi) =>
                `<option value="${vi}" ${vi === p.selectedVariant ? 'selected' : ''}>${v.name || (t('label_variant') + ' ' + (vi + 1))}</option>`
            ).join('');
            versionSelect = `
                <div class="program-variant">
                    <span class="label">${t('label_variant')}</span>
                    <select onchange="setProgramVersion(${idx}, this.value)" title="${t('title_select_variant')}">
                        ${opts}
                    </select>
                </div>
            `;
        }
        let installerPicker = '';
        if (p.installType === 'installer') {
            const current = p.installerFile || (p.variants ? '' : '(auto)');
            installerPicker = `
                <div class="program-installer-picker" data-idx="${idx}">
                    <span class="installer-label">${t('cfg_installer_label')}</span>
                    <span class="installer-current" title="${current}">${current}</span>
                    <button class="btn-mini" onclick="scanInstallerFiles(${idx})">📁 ${t('cfg_installer_pick')}</button>
                    <button class="btn-mini btn-reset" onclick="resetInstallerFile(${idx})" title="Reset">↺</button>
                </div>
            `;
        }
        return `
            <div class="program-item">
                <input type="checkbox" ${p.selected ? 'checked' : ''} onchange="toggleProgram(${idx}, this.checked)">
                <div class="info">
                    <div class="name">${p.name}</div>
                    <div class="desc">${p.description}</div>
                    ${versionSelect}
                    ${installerPicker}
                </div>
                <span class="badge ${installBadgeClass(p.installType)}">${p.installType}</span>
            </div>
        `;
    }).join('');

    updateCounts();
}

// Track the chosen variant per program (persists in memory on the program object).
function setProgramVersion(idx, versionIndex) {
    const p = allPrograms[idx];
    if (!p) return;
    p.selectedVariant = parseInt(versionIndex, 10) || 0;
    updateCounts();
}

// Scan the program's folder for installer candidates and let the user pick one.
async function scanInstallerFiles(idx) {
    const p = allPrograms[idx];
    if (!p) return;
    const folder = p.folderPath || '';
    if (!folder) { uiToast('Brak folderPath', { kind: 'warn' }); return; }

    const picker = document.querySelector(`.program-installer-picker[data-idx="${idx}"]`);
    if (!picker) return;

    // Show loading state
    let current = picker.querySelector('.installer-current');
    current.textContent = '...';

    try {
        const candidates = await window.api.scan_installer_candidates(folder);
        if (!candidates || candidates.length === 0) {
            current.textContent = '(brak plików)';
            return;
        }

        // Replace the picker with a dropdown
        const opts = candidates.map(c =>
            `<option value="${c.path}" ${c.path === (p.installerFile || '') ? 'selected' : ''}>${c.name}</option>`
        ).join('');
        picker.innerHTML = `
            <span class="installer-label">Instalator:</span>
            <select onchange="setInstallerFile(${idx}, this.value)" title="Wybierz plik instalatora">
                <option value="">(auto)</option>
                ${opts}
            </select>
            <button class="btn-mini btn-reset" onclick="resetInstallerFile(${idx})" title="Reset">↺</button>
        `;
        // Select current value
        const sel = picker.querySelector('select');
        if (p.installerFile) sel.value = p.installerFile;
    } catch (e) {
        current.textContent = '(błąd)';
    }
}

function setInstallerFile(idx, filePath) {
    const p = allPrograms[idx];
    if (!p) return;
    if (filePath) {
        p.installerFile = filePath;
    } else {
        delete p.installerFile;
    }
}

function resetInstallerFile(idx) {
    const p = allPrograms[idx];
    if (!p) return;
    delete p.installerFile;
    delete p._originalInstallerFile;
    renderPrograms();
}

function toggleProgram(idx, checked) {
    allPrograms[idx].selected = !!checked;
    updateCounts();
}

function updateCounts() {
    const count = allPrograms.filter(p => p.selected).length;
    document.getElementById('selected-count').textContent = count + ' ' + (typeof t === 'function' ? t('cfg_selected') : 'wybranych');
    document.getElementById('install-count').textContent = count;
}

function populateCategories() {
    const cats = [...new Set(allPrograms.map(p => p.category))].sort();
    const select = document.getElementById('filter-category');
    cats.forEach(c => {
        const opt = document.createElement('option');
        opt.value = c;
        opt.textContent = c;
        select.appendChild(opt);
    });
}

function initSearch() {
    document.getElementById('search').addEventListener('input', renderPrograms);
    document.getElementById('filter-type').addEventListener('change', renderPrograms);
    document.getElementById('filter-category').addEventListener('change', renderPrograms);
    document.getElementById('select-all').addEventListener('change', (e) => {
        const search = (document.getElementById('search').value || '').toLowerCase();
        const filterType = document.getElementById('filter-type').value;
        const filterCat = document.getElementById('filter-category').value;
        allPrograms.forEach(p => {
            let visible = true;
            if (search && !p.name.toLowerCase().includes(search)) visible = false;
            if (filterType !== 'all' && p.installType !== filterType) visible = false;
            if (filterCat !== 'all' && p.category !== filterCat) visible = false;
            if (visible) p.selected = e.target.checked;
        });
        renderPrograms();
    });

    const selectAllTweaks = document.getElementById('select-all-tweaks');
    if (selectAllTweaks) {
        selectAllTweaks.addEventListener('change', (e) => {
            allTweaks.forEach(t => {
                if (!t.removed && !t.situational) t.selected = e.target.checked;
            });
            renderTweaks();
        });
    }

    const selectAllDebloat = document.getElementById('select-all-debloat');
    if (selectAllDebloat) {
        selectAllDebloat.addEventListener('change', (e) => {
            allTweaks.forEach(t => {
                if (t.category === 'debloat' && !t.removed) t.selected = e.target.checked;
            });
            renderDebloat();
        });
    }
}
