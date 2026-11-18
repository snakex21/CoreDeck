// CoreDeck updater UI: the "update available" window, download progress,
// skip/later, and the post-update / rollback notices.
//
// Backend bindings (internal/api/api_update*.go): check_for_updates(mode),
// skip_update_version, start_update, update_progress, update_ready,
// get_update_settings, set_update_settings, get_rollback_info, rollback_update.
(function () {
    'use strict';

    const tt = (key, vars) => (typeof window.t === 'function' ? window.t(key, vars) : key);
    const api = () => window.api || {};
    const has = name => typeof api()[name] === 'function';

    function el(tag, className, text) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (text !== undefined) node.textContent = text;
        return node;
    }

    // Inline markdown subset of GitHub release notes: **bold**, `code`, [text](url).
    function appendInline(parent, text) {
        const pattern = /(\*\*[^*]+\*\*|`[^`]+`|\[[^\]]+\]\([^)]+\))/g;
        let last = 0;
        let m;
        while ((m = pattern.exec(text))) {
            if (m.index > last) parent.appendChild(document.createTextNode(text.slice(last, m.index)));
            const token = m[0];
            if (token.startsWith('**')) parent.appendChild(el('strong', '', token.slice(2, -2)));
            else if (token.startsWith('`')) parent.appendChild(el('code', '', token.slice(1, -1)));
            else parent.appendChild(document.createTextNode(token.replace(/^\[([^\]]+)\].*$/, '$1')));
            last = m.index + token.length;
        }
        if (last < text.length) parent.appendChild(document.createTextNode(text.slice(last)));
    }

    // Release notes (markdown) -> safe DOM: headings, bullet lists, paragraphs.
    // Text only; nothing from the release body is inserted as HTML.
    function renderNotes(markdown) {
        const root = el('div', 'cd-update-notes');
        let list = null;
        let paragraph = null;
        String(markdown || '').slice(0, 20000).split(/\r?\n/).forEach(raw => {
            const line = raw.trimEnd();
            const heading = line.match(/^#{1,6}\s+(.*)$/);
            const bullet = line.match(/^\s*[-*]\s+(.*)$/);
            if (!line.trim()) { list = null; paragraph = null; return; }
            if (heading) {
                list = null; paragraph = null;
                const h = el('h4');
                appendInline(h, heading[1].replace(/\*\*/g, ''));
                root.appendChild(h);
            } else if (bullet) {
                paragraph = null;
                if (!list) { list = el('ul'); root.appendChild(list); }
                const li = el('li');
                appendInline(li, bullet[1]);
                list.appendChild(li);
            } else if (list && /^\s{2,}\S/.test(raw)) {
                appendInline(list.lastChild, ' ' + line.trim());
            } else {
                list = null;
                if (!paragraph) { paragraph = el('p'); root.appendChild(paragraph); } else paragraph.appendChild(document.createTextNode(' '));
                appendInline(paragraph, line.trim());
            }
        });
        if (!root.childNodes.length) root.appendChild(el('p', '', '—'));
        return root;
    }

    function errorText(result) {
        switch (result && result.errorCode) {
            case 'bad_checksum': return tt('upd_bad_checksum');
            case 'bad_size': return tt('upd_bad_size');
            case 'no_asset': return tt('upd_no_asset');
            default: return tt('upd_failed', { error: (result && result.error) || '?' });
        }
    }

    // Shows the "update available" window. Resolves with "now" | "later" | "skip".
    async function offerUpdate(info) {
        const body = el('div', 'cd-update');
        const meta = el('div', 'cd-update-meta');
        meta.appendChild(el('span', 'cd-update-version', `CoreDeck ${info.latestVersion}`));
        if (info.prerelease) meta.appendChild(el('span', 'cd-update-badge', tt('upd_prerelease')));
        if (info.publishedAt) {
            const date = new Date(info.publishedAt);
            if (!Number.isNaN(date.getTime())) meta.appendChild(el('span', 'cd-update-date', date.toLocaleDateString(document.documentElement.lang || undefined)));
        }
        body.appendChild(meta);
        body.appendChild(el('h3', 'cd-update-heading', tt('upd_whats_new')));
        body.appendChild(renderNotes(info.notes));
        body.appendChild(el('p', 'cd-update-safety', tt('upd_safety_note')));

        const choice = await uiDialog({
            title: tt('upd_title'),
            message: tt('upd_message', { version: info.latestVersion, current: info.currentVersion }),
            body,
            wide: true,
            buttons: [
                { label: tt('upd_skip'), value: 'skip' },
                { label: tt('upd_later'), value: 'later' },
                { label: tt('upd_now'), value: 'now', variant: 'primary', confirm: true }
            ],
            cancelValue: 'later'
        });
        if (choice === 'skip') {
            if (has('skip_update_version')) await api().skip_update_version(info.latestVersion);
            uiToast(tt('upd_skipped', { version: info.latestVersion }), { kind: 'neutral' });
        } else if (choice === 'now') {
            await runUpdate(info.latestVersion);
        }
        return choice;
    }

    // Downloads, verifies and installs; shows progress in a toast.
    async function runUpdate(version) {
        if (!has('start_update')) return;
        const toast = uiToast(tt('upd_downloading', { percent: 0 }), { kind: 'info', title: `CoreDeck ${version}`, timeout: 0 });
        const start = await api().start_update(version);
        if (!start || start.error || !start.sessionId) {
            toast.update(errorText(start), 'error');
            return;
        }
        for (;;) {
            await new Promise(resolve => setTimeout(resolve, 300));
            const p = await api().update_progress(start.sessionId);
            if (!p || p.error && p.state !== 'error') { toast.update(errorText(p), 'error'); return; }
            if (p.state === 'downloading') toast.update(tt('upd_downloading', { percent: Math.max(0, Math.min(100, Math.round(p.percent || 0))) }));
            else if (p.state === 'verifying') toast.update(tt('upd_verifying'));
            else if (p.state === 'installing') toast.update(tt('upd_installing'));
            else if (p.state === 'restarting') { toast.update(tt('upd_restarting'), 'success'); return; }
            else if (p.state === 'error') { toast.update(errorText(p), 'error'); return; }
        }
    }

    // Manual check (Settings). Returns the backend result.
    async function checkNow() {
        if (!has('check_for_updates')) return { error: 'unsupported' };
        const result = await api().check_for_updates('manual');
        if (!result || result.error) return result || { error: '?' };
        if (result.updateAvailable) await offerUpdate(result);
        return result;
    }

    // On the start page: report a finished update or a rollback, then run the
    // automatic check (at most once a day; the backend decides).
    async function startupCheck() {
        if (has('update_ready')) {
            const ready = await api().update_ready();
            if (ready && ready.justUpdated) uiToast(tt('upd_done', { version: ready.justUpdated }), { kind: 'success', timeout: 6000 });
            if (ready && ready.rolledBack) {
                uiToast(tt('upd_rolled_back', { version: ready.rolledBack.to || '?', current: ready.rolledBack.from || '?' }), { kind: 'warn', timeout: 12000 });
            }
        }
        if (!has('check_for_updates')) return;
        await new Promise(resolve => setTimeout(resolve, 2500));
        const result = await api().check_for_updates('auto');
        if (result && !result.error && result.updateAvailable) await offerUpdate(result);
    }

    window.CoreUpdater = { offerUpdate, runUpdate, checkNow, startupCheck, renderNotes, errorText };

    if (document.body && document.body.dataset.updaterAutostart === 'true') {
        startupCheck().catch(e => console.warn('update check failed', e));
    }
})();
