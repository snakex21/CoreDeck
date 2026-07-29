// ============================================
// KNOWLEDGE BASE
// ============================================

let allKnowledge = [];
let allKnowledgeGroups = [];
let activeKnowledgeGroupId = '';
let activeKnowledgeCategoryIndex = -1;

async function loadKnowledgeData() {
    try {
        const data = await window.api.get_knowledge_data();
        allKnowledge = (data && data.categories) ? data.categories : [];
        allKnowledgeGroups = (data && data.groups) ? data.groups : [];
    } catch (e) {
        console.error('Knowledge load error:', e);
        allKnowledge = [];
        allKnowledgeGroups = [];
    }
    updateKnowledgeCount();
}

function updateKnowledgeCount() {
    let total = 0;
    allKnowledge.forEach(c => total += (c.links || []).length);
    const segCount = document.getElementById('seg-count-knowledge');
    if (segCount) segCount.textContent = total;
}

function renderKnowledgePane() {
    if (activeManageTab !== 'knowledge') return;
    renderKnowledgeList();
}

function renderKnowledgeList() {
    const list = document.getElementById('knowledge-list-manage');
    if (!list) return;

    const search = (document.getElementById('knowledge-search-manage')?.value || '').toLowerCase();
    if (!search && activeKnowledgeCategoryIndex < 0) {
        if (activeKnowledgeGroupId) renderKnowledgeCategories(list);
        else renderKnowledgeGroups(list);
        updateKnowledgeCount();
        return;
    }

    let total = 0;
    let html = '';
    const categoriesToRender = activeKnowledgeCategoryIndex >= 0 && !search
        ? [{ cat: allKnowledge[activeKnowledgeCategoryIndex], catIdx: activeKnowledgeCategoryIndex }]
        : allKnowledge.map((cat, catIdx) => ({ cat, catIdx }));

    if (activeKnowledgeCategoryIndex >= 0 && !search) {
        const cat = allKnowledge[activeKnowledgeCategoryIndex];
        html += `<div class="knowledge-detail-head"><button type="button" class="knowledge-detail-back" onclick="showKnowledgeCategories()">← Wszystkie kategorie</button><div class="knowledge-detail-title">${escapeHtml(cat?.icon || '📂')} ${escapeHtml(cat?.name || '')} <span class="knowledge-detail-count">${(cat?.links || []).length} linków</span></div></div>`;
    }

    categoriesToRender.forEach(({ cat, catIdx }) => {
        if (!cat) return;
        let links = cat.links || [];
        if (search) {
            links = links.filter(l =>
                (l.name || '').toLowerCase().includes(search) ||
                (l.url || '').toLowerCase().includes(search) ||
                (l.tag || '').toLowerCase().includes(search) ||
                (l.section || '').toLowerCase().includes(search)
            );
        }
        if (!links.length && search) return;
        total += links.length;

        // Group by section
        const sections = {};
        const noSection = [];
        links.forEach(l => {
            if (l.section) {
                if (!sections[l.section]) sections[l.section] = [];
                sections[l.section].push(l);
            } else {
                noSection.push(l);
            }
        });

        html += `<div style="padding:10px 24px 4px;font-size:var(--fs-2xs);font-weight:600;text-transform:uppercase;letter-spacing:0.05em;color:var(--text-muted);">${escapeHtml(cat.name)} <span style="color:var(--text-muted)">${links.length}</span></div>`;

        // Render sections
        Object.entries(sections).forEach(([sectionName, sectionLinks]) => {
            html += `<div style="padding:6px 24px 2px 36px;font-size:var(--fs-2xs);font-weight:500;color:var(--accent-primary);display:flex;align-items:center;gap:8px;">`;
            html += `📁 ${escapeHtml(sectionName)} <span style="color:var(--text-muted)">(${sectionLinks.length})</span>`;
            html += `<button class="row-action" style="width:20px;height:20px;" onclick="renameKnowledgeSection(${catIdx},'${escapeHtml(sectionName)}')" title="Zmień nazwę sekcji">${ICON.edit}</button>`;
            html += `<button class="row-action danger" style="width:20px;height:20px;" onclick="deleteKnowledgeSection(${catIdx},'${escapeHtml(sectionName)}')" title="Usuń sekcję (linki zostaną bez sekcji)">${ICON.trash}</button>`;
            html += `</div>`;
            sectionLinks.forEach(l => {
                const globalLinkIdx = (cat.links || []).indexOf(l);
                html += renderKnowledgeRow(l, catIdx, globalLinkIdx, true);
            });
        });

        // Render links without section
        if (noSection.length) {
            if (Object.keys(sections).length) {
                html += `<div style="padding:6px 24px 2px 36px;font-size:var(--fs-2xs);font-weight:500;color:var(--text-muted);">Inne</div>`;
            }
            noSection.forEach(l => {
                const globalLinkIdx = (cat.links || []).indexOf(l);
                html += renderKnowledgeRow(l, catIdx, globalLinkIdx, false);
            });
        }
    });

    if (!html) {
        list.innerHTML = `<div class="list-empty"><div class="list-empty-title">${t('empty_no_links')}</div><p style="font-size:var(--fs-xs);margin-top:4px;">${t('empty_add_first_link')}</p></div>`;
    } else {
        list.innerHTML = html;
    }
    updateKnowledgeCount();
}

function renderKnowledgeCategories(list) {
    const group = allKnowledgeGroups.find(item => item.id === activeKnowledgeGroupId);
    const categories = allKnowledge.map((category, index) => ({ category, index }))
        .filter(item => item.category.groupId === activeKnowledgeGroupId)
        .sort((a, b) => (a.category.order || 0) - (b.category.order || 0));
    if (!categories.length) {
        list.innerHTML = `<div class="knowledge-detail-head"><button type="button" class="knowledge-detail-back" onclick="showKnowledgeGroups()">← Wszystkie grupy</button><div class="knowledge-detail-title">${escapeHtml(group?.icon || '🗂️')} ${escapeHtml(group?.name || 'Grupa')}</div></div><div class="list-empty"><div class="list-empty-title">Ta grupa nie ma jeszcze kategorii</div><p style="font-size:var(--fs-xs);margin-top:4px;">Użyj przycisku „+ Kategoria”, aby dodać ją do tej grupy.</p></div>`;
        return;
    }
    list.innerHTML = `<div class="knowledge-detail-head"><button type="button" class="knowledge-detail-back" onclick="showKnowledgeGroups()">← Wszystkie grupy</button><div class="knowledge-detail-title">${escapeHtml(group?.icon || '🗂️')} ${escapeHtml(group?.name || 'Grupa')} <span class="knowledge-detail-count">${categories.length} kategorii</span></div></div><div class="knowledge-category-grid">${categories.map(({ category, index }) => {
        const links = category.links || [];
        const sections = new Set(links.map(link => link.section).filter(Boolean));
        const details = `${links.length} linków${sections.size ? ` · ${sections.size} sekcji` : ''}`;
        return `<button type="button" class="knowledge-category-card" onclick="showKnowledgeCategory(${index})">
            <span class="knowledge-category-icon">${escapeHtml(category.icon || '📂')}</span>
            <span><span class="knowledge-category-name">${escapeHtml(category.name || 'Bez nazwy')}</span><span class="knowledge-category-meta">${details}</span></span>
            <span class="knowledge-category-arrow">→</span>
        </button>`;
    }).join('')}</div>`;
}

function renderKnowledgeGroups(list) {
    if (!allKnowledgeGroups.length) {
        list.innerHTML = `<div class="list-empty"><div class="list-empty-title">Brak grup Bazy Wiedzy</div><p style="font-size:var(--fs-xs);margin-top:4px;">Dodaj grupy w danych Bazy Wiedzy.</p></div>`;
        return;
    }
    const groups = [...allKnowledgeGroups].sort((a, b) => (a.order || 0) - (b.order || 0));
    list.innerHTML = `<div class="knowledge-category-grid">${groups.map(group => {
        const categories = allKnowledge.filter(category => category.groupId === group.id);
        const linkCount = categories.reduce((total, category) => total + (category.links || []).length, 0);
        return `<button type="button" class="knowledge-category-card" onclick="showKnowledgeGroup('${escapeHtml(group.id)}')">
            <span class="knowledge-category-icon">${escapeHtml(group.icon || '🗂️')}</span>
            <span><span class="knowledge-category-name">${escapeHtml(group.name || 'Bez nazwy')}</span><span class="knowledge-category-meta">${categories.length} kategorii · ${linkCount} linków</span></span>
            <span class="knowledge-category-arrow">→</span>
        </button>`;
    }).join('')}</div>`;
}

function showKnowledgeGroup(groupId) {
    if (!allKnowledgeGroups.some(group => group.id === groupId)) return;
    activeKnowledgeGroupId = groupId;
    activeKnowledgeCategoryIndex = -1;
    const input = document.getElementById('knowledge-search-manage');
    if (input) input.value = '';
    renderKnowledgeList();
}

function showKnowledgeGroups() {
    activeKnowledgeGroupId = '';
    activeKnowledgeCategoryIndex = -1;
    const input = document.getElementById('knowledge-search-manage');
    if (input) input.value = '';
    renderKnowledgeList();
}

function showKnowledgeCategory(index) {
    if (!allKnowledge[index]) return;
    activeKnowledgeGroupId = allKnowledge[index].groupId || activeKnowledgeGroupId;
    activeKnowledgeCategoryIndex = index;
    const input = document.getElementById('knowledge-search-manage');
    if (input) input.value = '';
    renderKnowledgeList();
}

function showKnowledgeCategories() {
    activeKnowledgeCategoryIndex = -1;
    const input = document.getElementById('knowledge-search-manage');
    if (input) input.value = '';
    renderKnowledgeList();
}

function renderKnowledgeRow(l, catIdx, linkIdx, hasSection) {
    const tag = l.tag ? `<span class="badge badge-neutral" style="margin-left:var(--space-2);">${escapeHtml(l.tag)}</span>` : '';
    const star = l.starred ? '⭐ ' : '';
    const indent = hasSection ? 'padding-left:48px;' : '';
    return `
        <div class="row" style="${indent}">
            <div class="row-avatar">${l.starred ? '⭐' : '🔗'}</div>
            <div class="row-main">
                <div class="row-name">${star}${escapeHtml(l.name)}${tag}</div>
                <div class="row-meta">
                    <span>${escapeHtml(l.url)}</span>
                    ${l.note ? `<span class="row-meta-sep"></span><span>${escapeHtml(l.note)}</span>` : ''}
                </div>
            </div>
            <div class="row-actions">
                <button class="row-action" onclick="toggleKnowledgeStar(${catIdx},${linkIdx})" title="${l.starred ? t('title_unstar') : t('title_star')}" style="${l.starred ? 'color:var(--accent-primary);' : ''}">${l.starred ? '★' : '☆'}</button>
                <button class="row-action" onclick="window.api.open_url('${escapeHtml(l.url)}')" title="${t('title_open_browser')}">${ICON.folder}</button>
                <button class="row-action" onclick="editKnowledgeLink(${catIdx},${linkIdx})" title="${t('title_edit')}">${ICON.edit}</button>
                <button class="row-action danger" onclick="deleteKnowledgeLink(${catIdx},${linkIdx})" title="${t('title_delete')}">${ICON.trash}</button>
            </div>
        </div>
    `;
}

function openAddKnowledgeLink() {
    populateKnowledgeCategorySelect(activeKnowledgeCategoryIndex >= 0 ? activeKnowledgeCategoryIndex : undefined);
    const overlay = document.getElementById('knowledge-editor-overlay');
    if (!overlay) return;
    document.getElementById('knowledge-editor-title').textContent = 'Dodaj link';
    document.getElementById('kn-name').value = '';
    document.getElementById('kn-url').value = '';
    document.getElementById('kn-tag').value = '';
    document.getElementById('kn-section').value = '';
    document.getElementById('kn-note').value = '';
    document.getElementById('kn-starred').checked = false;
    delete overlay.dataset.editCat;
    delete overlay.dataset.editLink;
    overlay.classList.add('active');
}

function editKnowledgeLink(catIdx, linkIdx) {
    const cat = allKnowledge[catIdx];
    if (!cat || !cat.links || !cat.links[linkIdx]) return;
    const l = cat.links[linkIdx];
    populateKnowledgeCategorySelect(catIdx);
    const overlay = document.getElementById('knowledge-editor-overlay');
    if (!overlay) return;
    document.getElementById('knowledge-editor-title').textContent = 'Edytuj link';
    document.getElementById('kn-name').value = l.name || '';
    document.getElementById('kn-url').value = l.url || '';
    document.getElementById('kn-tag').value = l.tag || '';
    document.getElementById('kn-section').value = l.section || '';
    document.getElementById('kn-note').value = l.note || '';
    document.getElementById('kn-starred').checked = !!l.starred;
    overlay.dataset.editCat = catIdx;
    overlay.dataset.editLink = linkIdx;
    overlay.classList.add('active');
}

function closeKnowledgeEditor() {
    document.getElementById('knowledge-editor-overlay')?.classList.remove('active');
}

function populateKnowledgeCategorySelect(selectedIdx) {
    const sel = document.getElementById('kn-category');
    if (!sel) return;
    sel.innerHTML = '';
    allKnowledge.forEach((c, i) => {
        if (activeKnowledgeGroupId && c.groupId !== activeKnowledgeGroupId && i !== selectedIdx) return;
        const opt = document.createElement('option');
        opt.value = i;
        opt.textContent = c.name;
        if (i === selectedIdx) opt.selected = true;
        sel.appendChild(opt);
    });
    // Update section suggestions when category changes
    updateSectionSuggestions(selectedIdx);
}

function updateSectionSuggestions(catIdx) {
    const datalist = document.getElementById('kn-section-list');
    if (!datalist) return;
    datalist.innerHTML = '';
    const cat = allKnowledge[catIdx];
    if (!cat) return;
    const sections = new Set();
    (cat.links || []).forEach(l => {
        if (l.section) sections.add(l.section);
    });
    sections.forEach(s => {
        const opt = document.createElement('option');
        opt.value = s;
        datalist.appendChild(opt);
    });
}

async function saveKnowledgeLink() {
    const overlay = document.getElementById('knowledge-editor-overlay');
    if (!overlay) return;
    const name = document.getElementById('kn-name').value.trim();
    const url = document.getElementById('kn-url').value.trim();
    const catIdx = parseInt(document.getElementById('kn-category').value);
    const tag = document.getElementById('kn-tag').value.trim();
    const note = document.getElementById('kn-note').value.trim();
    const section = document.getElementById('kn-section').value.trim();
    const starred = document.getElementById('kn-starred').checked;

    if (!name) { showToast(t('val_name_req'), true); return; }
    if (!url) { showToast(t('val_url_req'), true); return; }
    if (isNaN(catIdx) || !allKnowledge[catIdx]) { showToast(t('val_select_cat'), true); return; }

    const editCat = overlay.dataset.editCat;
    const editLink = overlay.dataset.editLink;

    if (editCat !== undefined && editLink !== undefined) {
        allKnowledge[parseInt(editCat)].links.splice(parseInt(editLink), 1);
    }

    allKnowledge[catIdx].links.push({ name, url, tag, note, section, starred });
    closeKnowledgeEditor();
    await persistKnowledge();
    activeKnowledgeCategoryIndex = catIdx;
    renderKnowledgeList();
}

// Wire up category change to update section suggestions
document.addEventListener('DOMContentLoaded', () => {
    const catSel = document.getElementById('kn-category');
    if (catSel) {
        catSel.addEventListener('change', () => {
            updateSectionSuggestions(parseInt(catSel.value));
        });
    }
});

async function deleteKnowledgeLink(catIdx, linkIdx) {
    const cat = allKnowledge[catIdx];
    if (!cat || !cat.links || !cat.links[linkIdx]) return;
    const l = cat.links[linkIdx];
    if (!confirm(`Usunąć "${l.name}"?`)) return;
    cat.links.splice(linkIdx, 1);
    await persistKnowledge();
    renderKnowledgeList();
}

function openAddKnowledgeCategory() {
    if (!activeKnowledgeGroupId) {
        showToast('Najpierw otwórz grupę, do której chcesz dodać kategorię.', true);
        return;
    }
    const group = allKnowledgeGroups.find(item => item.id === activeKnowledgeGroupId);
    document.getElementById('kn-cat-name').value = '';
    const context = document.getElementById('kn-cat-group-context');
    if (context) context.textContent = `Nowa kategoria trafi do grupy: ${group?.icon || '🗂️'} ${group?.name || activeKnowledgeGroupId}`;
    document.getElementById('knowledge-cat-editor-overlay')?.classList.add('active');
}

function closeKnowledgeCategoryEditor() {
    document.getElementById('knowledge-cat-editor-overlay')?.classList.remove('active');
}

async function saveKnowledgeCategory() {
    const name = document.getElementById('kn-cat-name').value.trim();
    if (!name) { showToast(t('val_cat_name'), true); return; }
    const group = allKnowledgeGroups.find(item => item.id === activeKnowledgeGroupId);
    if (!group) { showToast('Nie znaleziono wybranej grupy.', true); return; }
    const usedIDs = new Set(allKnowledge.map(category => category.id));
    const baseID = slugify(name).replace(/_/g, '-') || 'kategoria';
    let id = baseID;
    let suffix = 2;
    while (usedIDs.has(id)) id = `${baseID}-${suffix++}`;
    const groupCategories = allKnowledge.filter(category => category.groupId === group.id);
    const order = groupCategories.reduce((max, category) => Math.max(max, Number(category.order) || 0), 0) + 1;
    allKnowledge.push({ id, groupId: group.id, name, icon: '📂', order, links: [] });
    closeKnowledgeCategoryEditor();
    await persistKnowledge();
    renderKnowledgeList();
}

async function persistKnowledge() {
    try {
        const result = await window.api.save_knowledge_data(JSON.stringify({
            version: '1.0',
            groups: allKnowledgeGroups,
            categories: allKnowledge
        }));
        if (result && result.error) {
            showToast(t('err_save_prefix') + result.error, true);
        } else {
            showToast(t('toast_saved'));
            updateKnowledgeCount();
        }
    } catch (e) {
        showToast(t('err_save_prefix') + e, true);
    }
}

async function renameKnowledgeSection(catIdx, oldName) {
    const cat = allKnowledge[catIdx];
    if (!cat) return;
    const newName = prompt(t('prompt_section_name'), oldName);
    if (!newName || newName === oldName) return;
    (cat.links || []).forEach(l => {
        if (l.section === oldName) l.section = newName;
    });
    await persistKnowledge();
    renderKnowledgeList();
}

async function deleteKnowledgeSection(catIdx, sectionName) {
    const cat = allKnowledge[catIdx];
    if (!cat) return;
    if (!confirm(`Usunąć sekcję "${sectionName}"? Linki zostaną ale bez sekcji.`)) return;
    (cat.links || []).forEach(l => {
        if (l.section === sectionName) l.section = '';
    });
    await persistKnowledge();
    renderKnowledgeList();
}

async function toggleKnowledgeStar(catIdx, linkIdx) {
    const cat = allKnowledge[catIdx];
    if (!cat || !cat.links || !cat.links[linkIdx]) return;
    cat.links[linkIdx].starred = !cat.links[linkIdx].starred;
    await persistKnowledge();
    renderKnowledgeList();
}
