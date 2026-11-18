let knowledgeData = { categories: [] };
let searchQuery = '';
let currentCategoryIndex = -1;
let currentGroupId = '';
const sectionOpenState = new Map();
const subsectionOpenState = new Map();

function getCatIcon(category) { return category?.icon || '📂'; }

function escapeHtml(s) {
    if (s == null) return '';
    return String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}

async function loadKnowledge() {
    try {
        const data = await window.api.get_knowledge_data();
        if (data && data.categories) {
            knowledgeData = data;
        }
    } catch (e) {
        console.error('Knowledge load error:', e);
    }
    renderAll();
    window.dispatchEvent(new CustomEvent('knowledge-data-loaded', { detail: knowledgeData }));
}

function renderStats() {
    const stats = document.getElementById('kb-stats');
    let total = 0;
    knowledgeData.categories.forEach(c => total += (c.links || []).length);
    const groupCount = (knowledgeData.groups || []).length;
    stats.innerHTML = `
        <div class="kb-stat"><div class="kb-stat-num">${groupCount}</div><div class="kb-stat-label">Grup</div></div>
        <div class="kb-stat"><div class="kb-stat-num">${knowledgeData.categories.length}</div><div class="kb-stat-label">${t('kb_stat_categories')}</div></div>
        <div class="kb-stat"><div class="kb-stat-num">${total}</div><div class="kb-stat-label">${t('kb_stat_links')}</div></div>
    `;
}

function getKnowledgeGroups() {
    const categoriesByGroup = new Map();
    knowledgeData.categories.forEach((category, index) => {
        if (!categoriesByGroup.has(category.groupId)) categoriesByGroup.set(category.groupId, []);
        categoriesByGroup.get(category.groupId).push({ category, index });
    });
    return (knowledgeData.groups || []).slice().sort((a, b) => (a.order || 0) - (b.order || 0)).map(group => {
        const categories = (categoriesByGroup.get(group.id) || []).sort((a, b) => (a.category.order || 0) - (b.category.order || 0));
        return { ...group, categories, links: categories.reduce((total, item) => total + (item.category.links || []).length, 0) };
    }).filter(group => group.categories.length);
}

function renderCategories() {
    const grid = document.getElementById('kb-cat-grid');
    const groupHeader = document.getElementById('kb-group-header');
    const empty = document.getElementById('kb-empty');
    let total = 0;
    knowledgeData.categories.forEach(c => total += (c.links || []).length);

    if (!total) {
        grid.innerHTML = '';
        groupHeader.innerHTML = '';
        groupHeader.className = '';
        grid.classList.remove('group-category-grid');
        empty.style.display = 'block';
        return;
    }
    empty.style.display = 'none';

    if (!currentGroupId) {
        groupHeader.innerHTML = '';
        groupHeader.className = '';
        grid.classList.remove('group-category-grid');
        grid.innerHTML = getKnowledgeGroups().map(group => {
            return `
            <button type="button" class="kb-cat-card" data-group-id="${escapeHtml(group.id)}" onclick="showGroup(this.dataset.groupId)">
                <div class="kb-cat-icon">${escapeHtml(group.icon)}</div>
                <div class="kb-cat-name">${escapeHtml(group.name)}</div>
                <div class="kb-cat-count">${group.categories.length} kategorii • ${group.links} ${t('kb_links_count')}</div>
            </button>`;
        }).join('');
        return;
    }

    const selectedGroup = getKnowledgeGroups().find(group => group.id === currentGroupId);
    if (!selectedGroup) {
        currentGroupId = '';
        renderCategories();
        return;
    }
    grid.classList.add('group-category-grid');
    groupHeader.className = 'kb-group-header';
    groupHeader.innerHTML = `<button class="kb-group-back" onclick="showCategories()">← Wszystkie grupy</button><div class="kb-group-icon">${escapeHtml(selectedGroup.icon)}</div><div class="kb-group-copy"><div class="kb-group-kicker">Grupa</div><h3>${escapeHtml(selectedGroup.name)}</h3><span>${selectedGroup.categories.length} kategorii • ${selectedGroup.links} ${t('kb_links_count')}</span></div>`;
    grid.innerHTML = selectedGroup.categories.map(({ category: cat, index: i }) => {
        const count = (cat.links || []).length;
        const icon = getCatIcon(cat);
        return `
            <button type="button" class="kb-cat-card" onclick="showCategory(${i})">
                <div class="kb-cat-icon">${icon}</div>
                <div class="kb-cat-name">${escapeHtml(cat.name)}</div>
                <div class="kb-cat-count">${count} ${t('kb_links_count')}</div>
            </button>
        `;
    }).join('');
}

function showGroup(groupId) {
    currentGroupId = groupId;
    currentCategoryIndex = -1;
    const group = getKnowledgeGroups().find(item => item.id === groupId);
    if (group?.categories.length === 1) {
        showCategory(group.categories[0].index);
        return;
    }
    document.getElementById('kb-detail-view').classList.remove('active');
    document.getElementById('kb-search-view').classList.remove('active');
    document.getElementById('kb-categories-view').style.display = 'block';
    renderCategories();
}

function showCategory(idx) {
    const cat = knowledgeData.categories[idx];
    if (!cat) return;
    hideReviewForOtherCategory(cat.name);
    currentCategoryIndex = idx;
    const links = cat.links || [];

    // Render tabs
    const tabsEl = document.getElementById('kb-tabs');
    tabsEl.innerHTML = knowledgeData.categories.map((c, i) => ({ c, i })).filter(({ c }) => !currentGroupId || c.groupId === currentGroupId).map(({ c, i }) => {
        const icon = `<span class="kb-tab-icon">${getCatIcon(c)}</span>`;
        return `<button class="kb-tab ${i === idx ? 'active' : ''}" onclick="showCategory(${i})">${icon}<span>${escapeHtml(c.name)}</span></button>`;
    }).join('');

    document.getElementById('kb-detail-title').innerHTML =
        `${getCatIcon(cat)} ${escapeHtml(cat.name)} <span style="color:var(--text-muted);font-size:0.9rem;font-weight:400;">(${links.length})</span>`;

    // Starred links are shown as shortcuts at the top, but remain in their normal sections too.
    const starred = links.filter(l => l.starred);
    const normal = links;

    const sectionTree = buildSectionTree(normal);

    let linksHtml = '';

    // Featured/starred links at top (always full-width list)
    if (starred.length) {
        linksHtml += `<div class="kb-featured-section">`;
        linksHtml += `<div class="kb-featured-title">⭐ ${t('kb_featured')}</div>`;
        linksHtml += `<div class="kb-links list-mode">${starred.map(l => renderLinkCard(l)).join('')}</div>`;
        linksHtml += `</div>`;
    }

    linksHtml += renderSectionTree(sectionTree);

    if (!linksHtml) {
        linksHtml = `<div class="kb-empty"><div class="kb-empty-icon">📭</div><div class="kb-empty-title">${t('kb_empty_no_links')}</div></div>`;
    }

    const linksEl = document.getElementById('kb-links');
    linksEl.innerHTML = linksHtml;
    linksEl.classList.toggle('list-mode', currentView === 'list');
    linksEl.classList.toggle('grid-mode', currentView === 'grid');
    updateViewButtons();

    document.getElementById('kb-categories-view').style.display = 'none';
    document.getElementById('kb-search-view').classList.remove('active');
    document.getElementById('kb-detail-view').classList.add('active');

}

function renderLinkCard(l) {
    const tag = l.tag ? `<span class="kb-link-tag">${escapeHtml(l.tag)}</span>` : '';
    const note = l.note ? `<div class="kb-link-note">${escapeHtml(l.note)}</div>` : '';
    const section = l.section ? `<span style="font-size:0.65rem;color:var(--text-muted);background:var(--bg-elevated);padding:2px 6px;border-radius:4px;">${escapeHtml(l.section)}</span>` : '';
    const star = l.starred ? `<span style="font-size:0.9rem;">⭐</span>` : '';
    const starClass = l.starred ? ' kb-link-starred' : '';
    return `
        <div class="kb-link${starClass}" role="button" tabindex="0" data-url="${escapeHtml(l.url)}" onclick="openKnowledgeUrl(this.dataset.url)" onkeydown="handleLinkKey(event, this)">
            <span class="kb-link-icon">${star || '🔗'}</span>
            <div class="kb-link-info">
                <div class="kb-link-name">${escapeHtml(l.name)}</div>
                <div class="kb-link-url">${escapeHtml(l.url)}</div>
                ${note}
            </div>
            ${section}
            ${tag}
        </div>
    `;
}

function parseSectionPath(section) {
    return String(section || '').split('/').map(s => s.trim()).filter(Boolean);
}

function buildSectionTree(links) {
    const root = { name: '', links: [], children: [], childMap: new Map() };
    links.forEach(link => {
        const parts = parseSectionPath(link.section);
        let node = root;
        parts.forEach(part => {
            if (!node.childMap.has(part)) {
                const child = { name: part, links: [], children: [], childMap: new Map() };
                node.childMap.set(part, child);
                node.children.push(child);
            }
            node = node.childMap.get(part);
        });
        node.links.push(link);
    });
    return root;
}

function countTreeLinks(node) {
    return node.links.length + node.children.reduce((total, child) => total + countTreeLinks(child), 0);
}

function treeKey(path) {
    return `${currentCategoryIndex}\u001f${path.join('\u001f')}`;
}

function renderSectionTree(root) {
    let html = '';
    if (root.links.length) {
        html += `<section class="kb-section-panel"><div class="kb-section-panel-head"><h2 class="kb-section-panel-title">${t('kb_other')}</h2><span class="kb-section-count">${root.links.length} ${t('kb_links_count')}</span></div>${renderCompactLinkList(root.links, treeKey(['root']))}</section>`;
    }
    html += root.children.map(node => renderSectionNode(node, [node.name], 1)).join('');
    return html;
}

function renderSectionNode(node, path, depth) {
    const key = treeKey(path);
    const open = sectionOpenState.has(key) ? sectionOpenState.get(key) : depth === 1;
    const count = countTreeLinks(node);
    const topLevel = depth === 1;
    const panelClass = topLevel ? 'kb-section-panel' : 'kb-subsection-panel';
    const summaryClass = topLevel ? 'kb-section-panel-head' : 'kb-subsection-title-row';
    const title = topLevel
        ? `<h2 class="kb-section-panel-title">${escapeHtml(node.name)}</h2>`
        : `<h3 class="kb-subsection-title">${escapeHtml(node.name)}</h3>`;
    let content = '';
    if (open) {
        if (node.links.length) content += renderCompactLinkList(node.links, `${key}\u001flinks`);
        if (node.children.length) {
            const children = node.children.map(child => renderSectionNode(child, [...path, child.name], depth + 1)).join('');
            content += topLevel ? `<div class="kb-subsection-grid">${children}</div>` : `<div class="kb-nested-tree">${children}</div>`;
        }
    }
    return `<details class="${panelClass}" ${open ? 'open' : ''}>
        <summary class="${summaryClass}" data-tree-key="${escapeHtml(key)}" onclick="toggleSectionNode(event, this)">
            ${title}<span class="kb-section-count">${count} ${topLevel ? t('kb_links_count') : ''}</span>
        </summary>${content}</details>`;
}

function toggleSectionNode(event, element) {
    event.preventDefault();
    const key = element.dataset.treeKey;
    if (!key) return;
    sectionOpenState.set(key, !element.parentElement.open);
    showCategory(currentCategoryIndex);
}

function renderCompactLinkList(links) {
    return `<ul class="kb-compact-list">${links.map(renderCompactLink).join('')}</ul>`;
}

function renderCompactLink(l) {
    const tag = l.tag ? `<span class="kb-compact-tag">${escapeHtml(l.tag)}</span>` : '';
    const note = l.note ? `<div class="kb-compact-note">${escapeHtml(l.note)}</div>` : '';
    const star = l.starred ? `<span class="kb-compact-star">⭐</span>` : '';
    const starClass = l.starred ? ' kb-compact-starred' : '';
    return `
        <li class="kb-compact-item${starClass}">
            <div class="kb-compact-main">
                <a class="kb-compact-link" href="${escapeHtml(l.url)}" onclick="event.preventDefault(); openKnowledgeUrl(this.href)">${star}${escapeHtml(l.name)}</a>
                <div class="kb-compact-url">${escapeHtml(l.url)}</div>
            </div>
            ${tag}
            ${note}
        </li>
    `;
}

let currentView = localStorage.getItem('kb-view') || 'list';

if (!['list', 'grid'].includes(currentView)) currentView = 'list';

function updateViewButtons() {
    const listBtn = document.getElementById('kb-view-list');
    const gridBtn = document.getElementById('kb-view-grid');
    if (!listBtn || !gridBtn) return;
    listBtn.classList.toggle('active', currentView === 'list');
    gridBtn.classList.toggle('active', currentView === 'grid');
}

function setView(mode) {
    if (!['list', 'grid'].includes(mode)) return;
    currentView = mode;
    localStorage.setItem('kb-view', mode);
    updateViewButtons();
    if (currentCategoryIndex >= 0) {
        showCategory(currentCategoryIndex);
    }
}

function initViewToggle() {
    updateViewButtons();
}

function openKnowledgeUrl(url) {
    if (!url) return;
    window.api.open_url(url);
}

function handleLinkKey(event, el) {
    if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault();
        openKnowledgeUrl(el.dataset.url);
    }
}

let linkReviewSessionId = '';
let linkReviewCategory = '';
let linkReviewData = null;
let linkReviewTimer = null;
let linkReviewPage = 0;
const REVIEW_PAGE_SIZE = 10;
const linkReviewDecisions = new Map();
const linkReviewWatchers = new Set();

function reviewKey(result) {
    return `${result.category}\n${result.url}`;
}

function reviewStateIcon(state) {
    return state === 'pending' ? '⏳' : state === 'ok' ? '✅' : state === 'warning' ? '⚠️' : '❌';
}

function hideReviewForOtherCategory(categoryName) {
    if (!linkReviewCategory || linkReviewCategory === categoryName) return;
    clearTimeout(linkReviewTimer);
    linkReviewTimer = null;
    linkReviewSessionId = '';
    linkReviewCategory = '';
    linkReviewData = null;
    linkReviewPage = 0;
    const panel = document.getElementById('kb-review-panel');
    panel.classList.remove('active');
    panel.innerHTML = '';
}

async function startCategoryReview() {
    const category = knowledgeData.categories[currentCategoryIndex];
    if (!category) return;
    if (!window.api || !window.api.start_knowledge_link_check) {
        uiToast('Moduł weryfikacji linków wymaga ponownego uruchomienia launchera.', { kind: 'error' });
        return;
    }
    clearTimeout(linkReviewTimer);
    linkReviewCategory = category.name;
    const panel = document.getElementById('kb-review-panel');
    panel.classList.add('active');
    panel.innerHTML = `<div class="kb-review-head"><h3>🩺 Weryfikacja: ${escapeHtml(category.name)}</h3><span class="kb-review-progress">Przygotowywanie kolejki…</span></div><p class="kb-review-note">Przygotowywana jest stała kolejka do 50 adresów. Wynik techniczny nie usuwa żadnego linku — decyzję podejmujesz przyciskiem ✓, ⚠ lub ✕.</p>`;
    const started = await window.api.start_knowledge_link_check(category.name, 50);
    if (!started || started.error) {
        panel.innerHTML = `<div class="kb-review-head"><h3>🩺 Weryfikacja linków</h3></div><p class="kb-review-note">${escapeHtml(started?.error || 'Nie udało się uruchomić sprawdzania.')}</p>`;
        return;
    }
    linkReviewSessionId = started.sessionId;
    linkReviewPage = 0;
    linkReviewData = { total: started.total || 0, checked: 0, results: [], done: false };
    refreshLinkReview();
}

async function refreshLinkReview() {
    if (!linkReviewSessionId) return;
    const activeCategory = knowledgeData.categories[currentCategoryIndex];
    if (!activeCategory || activeCategory.name !== linkReviewCategory) return;
    const data = await window.api.get_knowledge_link_check(linkReviewSessionId);
    if (!data || data.error) {
        document.getElementById('kb-review-panel').innerHTML = `<p class="kb-review-note">${escapeHtml(data?.error || 'Weryfikacja została przerwana.')}</p>`;
        return;
    }
    linkReviewData = data;
    renderLinkReview();
    if (!data.done) {
        linkReviewTimer = setTimeout(refreshLinkReview, 650);
    }
}

function renderLinkReview() {
    const panel = document.getElementById('kb-review-panel');
    if (!linkReviewData) return;
    const results = linkReviewData.results || [];
    const progress = `${linkReviewData.checked || 0} / ${linkReviewData.total || 0}${linkReviewData.done ? ' • gotowe' : ' • sprawdzanie…'}`;
    const pageCount = Math.max(1, Math.ceil(results.length / REVIEW_PAGE_SIZE));
    if (linkReviewPage >= pageCount) linkReviewPage = pageCount - 1;
    const pageStart = linkReviewPage * REVIEW_PAGE_SIZE;
    const pageResults = results.slice(pageStart, pageStart + REVIEW_PAGE_SIZE);
    const isUnreviewedGreen = result => result.state === 'ok' && !linkReviewDecisions.has(reviewKey(result));
    const pageGreen = pageResults.filter(isUnreviewedGreen).length;
    const allGreen = results.filter(isUnreviewedGreen).length;
    const items = pageResults.map((result, pageIndex) => {
        const index = pageStart + pageIndex;
        const decision = linkReviewDecisions.get(reviewKey(result));
        const status = result.statusCode ? `HTTP ${result.statusCode} · ` : '';
        const actions = decision
            ? `<span class="kb-review-decision">${decision}</span>`
            : result.state === 'pending'
                ? `<span class="kb-review-progress">⏳ w kolejce</span>`
            : `<button class="kb-review-action" title="Otwórz w oknie kontroli" data-category="${escapeHtml(result.category)}" data-url="${escapeHtml(result.url)}" onclick="openLinkReview(this.dataset.category, this.dataset.url)">↗</button>
               <button class="kb-review-action" title="Zostaw link" onclick="reviewLink(${index}, 'keep')">✓</button>
               <button class="kb-review-action" title="Oznacz do późniejszej kontroli" onclick="reviewLink(${index}, 'flag')">⚠</button>
               <button class="kb-review-action remove" title="Usuń link" onclick="reviewLink(${index}, 'remove')">✕</button>`;
        return `<article class="kb-review-item">
            <span class="kb-review-state">${reviewStateIcon(result.state)}</span>
            <div style="min-width:0">
                <div class="kb-review-name">${escapeHtml(result.name)}${result.starred ? ' ⭐' : ''}</div>
                <div class="kb-review-url">${escapeHtml(result.url)}</div>
                <div class="kb-review-detail">${escapeHtml(status + (result.detail || ''))}</div>
            </div>
            <div class="kb-review-actions">${actions}</div>
        </article>`;
    }).join('');
    panel.classList.add('active');
    panel.innerHTML = `<div class="kb-review-head"><h3>🩺 Weryfikacja linków</h3><span class="kb-review-progress">${progress}</span></div>
        <p class="kb-review-note">Kolejka ma stały porządek: skaner zmienia tylko stan ⏳ na wynik, więc żadna pozycja nie przeskoczy na inną stronę. ✓ zostawia link, ⚠ oznacza go, a ✕ usuwa po potwierdzeniu.</p>
        <div class="kb-review-bulk">
            <button class="kb-review-bulk-btn" onclick="bulkKeepGreen('page')" ${pageGreen ? '' : 'disabled'}>✓ Zatwierdź zielone na tej stronie (${pageGreen})</button>
            <button class="kb-review-bulk-btn" onclick="bulkKeepGreen('all')" ${allGreen ? '' : 'disabled'}>✓ Zatwierdź wszystkie zielone (${allGreen})</button>
        </div>
        <div class="kb-review-list">${items || '<div class="kb-review-note">Czekam na kolejkę…</div>'}</div>
        <div class="kb-review-pagination">
            <button class="kb-review-page-btn" onclick="setLinkReviewPage(${linkReviewPage - 1})" ${linkReviewPage === 0 ? 'disabled' : ''}>← Poprzednie</button>
            <span class="kb-review-page-info">Strona ${linkReviewPage + 1} z ${pageCount} • pozycje ${results.length ? pageStart + 1 : 0}–${Math.min(pageStart + REVIEW_PAGE_SIZE, results.length)}</span>
            <button class="kb-review-page-btn" onclick="setLinkReviewPage(${linkReviewPage + 1})" ${linkReviewPage + 1 >= pageCount ? 'disabled' : ''}>Następne →</button>
        </div>`;
}

function setLinkReviewPage(page) {
    const total = Math.max(1, Math.ceil((linkReviewData?.results?.length || 0) / REVIEW_PAGE_SIZE));
    if (page < 0 || page >= total) return;
    linkReviewPage = page;
    renderLinkReview();
}

async function bulkKeepGreen(scope) {
    const results = linkReviewData?.results || [];
    const pageStart = linkReviewPage * REVIEW_PAGE_SIZE;
    const pool = scope === 'page' ? results.slice(pageStart, pageStart + REVIEW_PAGE_SIZE) : results;
    const candidates = pool.filter(result => result.state === 'ok' && !linkReviewDecisions.has(reviewKey(result)));
    if (!candidates.length) return;
    const label = scope === 'page' ? 'na tej stronie' : 'w całej obecnie zeskanowanej kolejce';
    if (!(await uiConfirm(`Zatwierdzić ${candidates.length} technicznie działających linków ${label}?\n\nNie obejrzysz ich ręcznie, ale decyzja ✓ zostanie zapisana.`, { confirmLabel: 'Zatwierdź' }))) return;
    const response = await window.api.bulk_keep_knowledge_links(JSON.stringify({
        category: candidates[0].category,
        urls: candidates.map(result => result.url),
        confirmed: true,
    }));
    if (!response || response.error) {
        uiToast(response?.error || 'Nie udało się zatwierdzić linków.', { kind: 'error' });
        return;
    }
    candidates.forEach(result => linkReviewDecisions.set(reviewKey(result), '✓ zatwierdzono zbiorczo'));
    renderLinkReview();
}

async function reviewLink(index, action) {
    const result = linkReviewData?.results?.[index];
    if (!result) return;
    let confirmed = false;
    if (action === 'remove') {
        const protection = result.starred ? '\n\nTo jest wyróżniony link — czy na pewno chcesz go usunąć?' : '';
        confirmed = await uiConfirm(`Usunąć „${result.name}” z kategorii „${result.category}”?${protection}`, { danger: true });
        if (!confirmed) return;
    }
    const response = await window.api.review_knowledge_link(JSON.stringify({
        action, category: result.category, url: result.url, confirmed,
    }));
    if (!response || response.error) {
        uiToast(response?.error || 'Nie udało się zapisać decyzji.', { kind: 'error' });
        return;
    }
    linkReviewDecisions.set(reviewKey(result), action === 'keep' ? '✓ zostawiono' : action === 'flag' ? '⚠ do sprawdzenia' : '✕ usunięto');
    if (action === 'remove') {
        await loadKnowledge();
    }
    renderLinkReview();
}

async function openLinkReview(category, url) {
    if (!window.api || !window.api.open_knowledge_link_review) {
        openKnowledgeUrl(url);
        return;
    }
    const response = await window.api.open_knowledge_link_review(category, url);
    if (!response || response.error) {
        uiToast(response?.error || 'Nie udało się otworzyć okna kontroli.', { kind: 'error' });
        return;
    }
    watchExternalLinkReview(category, url, response.openedAt);
}

function reviewDecisionText(action) {
    return action === 'keep' ? '✓ zostawiono' : action === 'flag' ? '⚠ do sprawdzenia' : '✕ usunięto';
}

async function watchExternalLinkReview(category, url, openedAt) {
    const key = reviewKey({ category, url });
    if (linkReviewWatchers.has(key) || !window.api?.get_knowledge_link_review_decision) return;
    linkReviewWatchers.add(key);
    const openedAtMs = Date.parse(openedAt || '') || Date.now();
    try {
        // The review window is separate from the launcher, so use its
        // persisted decision as the source of truth when it closes.
        for (let attempt = 0; attempt < 90; attempt++) {
            await new Promise(resolve => setTimeout(resolve, 400));
            const response = await window.api.get_knowledge_link_review_decision(category, url);
            const decision = response?.decision;
            const decisionAtMs = Date.parse(decision?.updatedAt || '');
            if (!response?.found || !decision?.action || !Number.isFinite(decisionAtMs) || decisionAtMs < openedAtMs) continue;

            linkReviewDecisions.set(key, reviewDecisionText(decision.action));
            if (decision.action === 'remove') {
                await loadKnowledge();
                if (currentCategoryIndex >= 0) showCategory(currentCategoryIndex);
            }
            renderLinkReview();
            return;
        }
    } finally {
        linkReviewWatchers.delete(key);
    }
}

function showCategories() {
    document.getElementById('kb-search').value = '';
    searchQuery = '';
    currentCategoryIndex = -1;
    currentGroupId = '';
    document.getElementById('kb-detail-view').classList.remove('active');
    document.getElementById('kb-search-view').classList.remove('active');
    document.getElementById('kb-categories-view').style.display = 'block';
    renderCategories();
}

function backFromCategory() {
    if (currentGroupId) {
        const group = getKnowledgeGroups().find(item => item.id === currentGroupId);
        if (group?.categories.length === 1) {
            showCategories();
            return;
        }
        showGroup(currentGroupId);
        return;
    }
    showCategories();
}

function renderSearch() {
    const view = document.getElementById('kb-search-view');
    const content = document.getElementById('kb-search-content');
    const q = searchQuery.toLowerCase().trim();

    if (!q) {
        view.classList.remove('active');
        return;
    }

    const results = [];
    const SEARCH_RENDER_LIMIT = 150;
    let totalMatches = 0;
    knowledgeData.categories.forEach(cat => {
        (cat.links || []).forEach(l => {
            if ((l.name || '').toLowerCase().includes(q) ||
                (l.url || '').toLowerCase().includes(q) ||
                (l.tag || '').toLowerCase().includes(q) ||
                (l.note || '').toLowerCase().includes(q) ||
                (l.section || '').toLowerCase().includes(q) ||
                (cat.name || '').toLowerCase().includes(q)) {
                totalMatches++;
                if (results.length < SEARCH_RENDER_LIMIT) {
                    results.push({ ...l, categoryId: cat.id, categoryName: cat.name, categoryIcon: getCatIcon(cat) });
                }
            }
        });
    });

    view.classList.add('active');
    document.getElementById('kb-categories-view').style.display = 'none';
    document.getElementById('kb-detail-view').classList.remove('active');

    if (!results.length) {
        content.innerHTML = `<div class="kb-empty"><div class="kb-empty-icon">🔍</div><div class="kb-empty-title">${t('no_ai_results')} "${escapeHtml(searchQuery)}"</div></div>`;
        return;
    }

    // Group by data-defined category identity.
    const grouped = {};
    results.forEach(r => {
        if (!grouped[r.categoryId]) grouped[r.categoryId] = { name: r.categoryName, icon: r.categoryIcon, links: [] };
        grouped[r.categoryId].links.push(r);
    });

    const limitNote = totalMatches > results.length ? ` • pokazuję pierwsze ${results.length}; doprecyzuj wyszukiwanie` : '';
    content.innerHTML = `<div class="kb-search-section-title">${totalMatches} ${t('kb_results_count')} "${escapeHtml(searchQuery)}"${limitNote}</div>` +
        Object.values(grouped).map(category => {
            const linksHtml = category.links.map(l => {
                const tag = l.tag ? `<span class="kb-link-tag">${escapeHtml(l.tag)}</span>` : '';
                return `
                    <div class="kb-link" role="button" tabindex="0" data-url="${escapeHtml(l.url)}" onclick="openKnowledgeUrl(this.dataset.url)" onkeydown="handleLinkKey(event, this)">
                        <span class="kb-link-icon">🔗</span>
                        <div class="kb-link-info">
                            <div class="kb-link-name">${escapeHtml(l.name)}</div>
                            <div class="kb-link-url">${escapeHtml(l.url)}</div>
                        </div>
                        ${tag}
                    </div>
                `;
            }).join('');
            return `
                <div class="kb-search-section">
                    <div class="kb-search-section-title">${escapeHtml(category.icon)} ${escapeHtml(category.name)} (${category.links.length})</div>
                    <div class="kb-links">${linksHtml}</div>
                </div>
            `;
        }).join('');
}

function renderAll() {
    renderStats();
    renderCategories();
}

// Search
const searchInput = document.getElementById('kb-search');
let searchTimer = null;
searchInput.addEventListener('input', (e) => {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => {
        searchQuery = e.target.value;
        if (searchQuery.trim()) {
            renderSearch();
        } else {
            document.getElementById('kb-search-view').classList.remove('active');
            document.getElementById('kb-categories-view').style.display = 'block';
        }
    }, 150);
});

// Theme
function initTheme() {
    const btn = document.getElementById('theme-toggle-btn');
    const icon = btn.querySelector('.theme-icon');
    if (localStorage.getItem('launcher-theme') === 'light') {
        document.body.classList.add('light-theme');
        if (icon) icon.textContent = '☀️';
    }
    btn.addEventListener('click', () => {
        document.body.classList.toggle('light-theme');
        const light = document.body.classList.contains('light-theme');
        localStorage.setItem('launcher-theme', light ? 'light' : 'dark');
        if (icon) icon.textContent = light ? '☀️' : '🌙';
    });
}

initTheme();
initViewToggle();
loadKnowledge();

// === i18n init for subpage ===
if (typeof applyTranslations === 'function') applyTranslations();
const langSelect = document.getElementById('lang-select');
if (langSelect && typeof getCurrentLang === 'function') {
    langSelect.value = getCurrentLang();
    langSelect.addEventListener('change', (e) => {
        if (typeof setLanguage === 'function') setLanguage(e.target.value);
    });
}
