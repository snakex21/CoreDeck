// ============================================
// WŁASNE STRONY (custom HTML pages)
// Strony są programami z type:"link" wskazującymi na plik HTML.
// ============================================

function renderPagesPane() {
    if (activeManageTab !== 'pages') return;
    renderPagesList();
}

function getPages() {
    const term = (document.getElementById('pages-search-manage')?.value || '').toLowerCase().trim();
    let pages = allPrograms.filter(p => p.type === 'link');
    if (term) {
        pages = pages.filter(p =>
            (p.name || '').toLowerCase().includes(term) ||
            (p.description || '').toLowerCase().includes(term) ||
            (p.category || '').toLowerCase().includes(term) ||
            (p.localPath || '').toLowerCase().includes(term)
        );
    }
    return pages;
}

function renderPagesList() {
    const list = document.getElementById('pages-list-manage');
    if (!list) return;
    const pages = getPages();

    if (!pages.length) {
        list.innerHTML = `
            <div class="list-empty">
                ${ICON.inbox}
                <div class="list-empty-title">${t('empty_no_pages')}</div>
                <div>${t('empty_create_page')}</div>
            </div>`;
        return;
    }

    list.innerHTML = pages.map((p, idx) => {
        const iconHtml = renderProgramIcon(p);
        const localPath = escapeHtml(p.localPath || '');
        const encodedPath = encodeURIComponent(p.localPath || '');
        const encodedUrl = encodeURIComponent(p.externalUrl || '');
        const encodedName = encodeURIComponent(p.name || '');
        const name = escapeHtml(p.name);
        const cat = escapeHtml(p.category || t('default_pages_cat'));
        const desc = escapeHtml(p.description || '');
        return `
            <div class="row">
                <div class="row-avatar">${iconHtml}</div>
                <div class="row-main">
                    <div class="row-name">${name}</div>
                    <div class="row-meta">
                        <span class="${badgeClass('page')}">${t('chip_page_label')}</span>
                        <span>${cat}</span>
                        ${desc ? `<span class="row-meta-sep"></span><span>${desc}</span>` : ''}
                    </div>
                    <div class="row-meta" style="margin-top:1px;">
                        <span style="font-family:ui-monospace,'SF Mono',Menlo,Consolas,monospace;font-size:var(--fs-2xs);color:var(--text-muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${localPath}</span>
                    </div>
                </div>
                <div class="row-actions" style="opacity:1;">
                    <button class="row-action" title="${p.localPath ? t('title_preview_page') : t('title_open_browser')}" onclick="previewPage('${encodedPath}', true, '${encodedUrl}')">${ICON.inbox}</button>
                    <button class="row-action" title="${t('title_edit_entry')}" onclick="openEditProgram('${encodedName}')">${ICON.edit}</button>
                    <button class="row-action" title="${t('title_open_html')}" onclick="openPageFolder('${encodedPath}', true)">${ICON.folder}</button>
                    <button class="row-action danger" title="${t('title_delete_page')}" onclick="confirmDeletePage(${idx})">${ICON.trash}</button>
                </div>
            </div>`;
    }).join('');
}

function openAddPage() {
    document.getElementById('page-name').value = '';
    document.getElementById('page-desc').value = '';
    document.getElementById('page-category').value = '';
    document.getElementById('page-icon').value = '';

    const suggest = document.getElementById('page-category-suggest');
    if (suggest) {
        const cats = [...new Set(allPrograms.map(p => p.category).filter(Boolean))].sort();
        suggest.innerHTML = cats.map(c => `<option value="${escapeHtml(c)}">`).join('');
    }

    renderPageIconPreview('');
    document.getElementById('page-editor-overlay').classList.add('active');
    setTimeout(() => document.getElementById('page-name')?.focus(), 50);
}

function closePageEditor() {
    document.getElementById('page-editor-overlay')?.classList.remove('active');
}

async function savePage() {
    const name = document.getElementById('page-name').value.trim();
    const description = document.getElementById('page-desc').value.trim();
    const category = document.getElementById('page-category').value.trim() || t('default_pages_cat');
    const icon = document.getElementById('page-icon').value.trim();

    if (!name) { showToast(t('val_page_name'), true); return; }

    try {
        const result = await window.api.create_page(JSON.stringify({ name, description, category, icon }));
        if (result && result.error) {
            showToast(t('upv_create_error', { error: tErr(result) }), true);
            return;
        }
        showToast(t('upv_page_created', { name }));
        closePageEditor();
        await loadPrograms();
        renderPagesList();
    } catch (e) {
        showToast(t('upv_create_error', { error: String(e) }), true);
    }
}

async function confirmDeletePage(pageListIndex) {
    const page = getPages()[pageListIndex];
    if (!page) return;
    if (!(await uiConfirm(t('upv_delete_confirm', { name: page.name }), { danger: true }))) return;
    deletePage(page);
}

async function deletePage(page) {
    try {
        const result = await window.api.delete_page(JSON.stringify({ name: page.name, localPath: page.localPath || '' }));
        if (result && result.error) {
            showToast(t('upv_delete_error', { error: tErr(result) }), true);
            return;
        }
        showToast(t('toast_page_deleted'));
        await loadPrograms();
        renderPagesList();
    } catch (e) {
        showToast(t('upv_delete_error', { error: String(e) }), true);
    }
}

function openPageFolder(localPath, encoded = false) {
    if (encoded) localPath = decodeURIComponent(localPath || '');
    if (!localPath) return;
    const slash = localPath.lastIndexOf('/');
    const folderPath = slash >= 0 ? localPath.slice(0, slash) : localPath;
    window.api.open_folder(folderPath || localPath);
}

function previewPage(localPath, encoded = false, externalUrl = '') {
    if (encoded) {
        localPath = decodeURIComponent(localPath || '');
        externalUrl = decodeURIComponent(externalUrl || '');
    }
    if (!localPath) {
        // Wpisy type:"link" bez localPath to zwykłe linki zewnętrzne (externalUrl).
        // Nie ma pliku do pokazania w iframe — otwieramy je w przeglądarce systemowej,
        // żeby przycisk podglądu nie był martwy.
        if (externalUrl) { window.api.open_url(externalUrl); return; }
        showToast(t('upv_preview_no_target'), true);
        return;
    }
    const overlay = document.getElementById('page-preview-overlay');
    const iframe = document.getElementById('page-preview-frame');
    const title = document.getElementById('page-preview-title');
    if (!overlay || !iframe) { window.open(localPath, '_self'); return; }
    title.textContent = decodeURIComponent(localPath.split('/').pop() || localPath);
    iframe.src = localPath;
    overlay.classList.add('active');
}

function closePagePreview() {
    const overlay = document.getElementById('page-preview-overlay');
    const iframe = document.getElementById('page-preview-frame');
    if (overlay) overlay.classList.remove('active');
    if (iframe) iframe.src = 'about:blank';
}

function pickPageIcon() {
    document.getElementById('page-icon-file')?.click();
}

function renderPageIconPreview(iconValue) {
    const preview = document.getElementById('page-icon-preview');
    if (!preview) return;
    const value = iconValue || '';
    if (!value) {
        preview.innerHTML = '📄';
        return;
    }
    if (value.startsWith('data:image/') || isPathIcon(value)) {
        preview.innerHTML = `<img src="${escapeHtml(value)}" alt="" onerror="this.outerHTML='📄'" />`;
        return;
    }
    preview.textContent = value;
}

function initPageIconPicker() {
    const fileInput = document.getElementById('page-icon-file');
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
            document.getElementById('page-icon').value = dataUrl;
            renderPageIconPreview(dataUrl);
        };
        reader.onerror = () => showToast(t('error_image_load'), true);
        reader.readAsDataURL(file);
    });
}
