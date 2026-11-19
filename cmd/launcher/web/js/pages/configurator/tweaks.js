// --- Tweaks ---
async function loadTweaks() {
    try {
        const data = await window.api.get_tweaks_data();
        if (data.error) {
            throw new Error(data.error);
        }
        let tweaks = data.tweaks || [];
        const isWin11 = systemInfo && systemInfo.isWin11;
        const isLaptop = systemInfo && systemInfo.isLaptop;
        const hasAmdGpu = systemInfo && systemInfo.hasAmdGpu;
        const hasNvidiaGpu = systemInfo && systemInfo.hasNvidiaGpu;
        const hasIntelGpu = systemInfo && systemInfo.hasIntelGpu;
        allTweaks = tweaks.filter(t => {
            if (t.os && t.os !== 'all') {
                if (t.os === 'win11' && !isWin11) return false;
                if (t.os === 'win10' && isWin11) return false;
            }
            if (t.device && t.device !== 'all') {
                if (t.device === 'laptop' && !isLaptop) return false;
                if (t.device === 'desktop' && isLaptop) return false;
            }
            if (t.gpu === 'amd' && !hasAmdGpu) return false;
            if (t.gpu === 'nvidia' && !hasNvidiaGpu) return false;
            if (t.gpu === 'intel' && !hasIntelGpu) return false;
            return true;
        });
        renderTweaks();
    } catch (e) {
        console.error('Failed to load tweaks:', e);
        const list = document.getElementById('tweaks-list');
        if (list) list.innerHTML = `<div class="program-item"><div class="info"><div class="name" style="color:var(--danger)">${t('err_tweaks_load')}</div><div class="desc">${String(e)}</div></div></div>`;
    }
}

function renderTweaks() {
    const list = document.getElementById('tweaks-list');
    if (!list) return;

    if (!allTweaks.length) {
        list.innerHTML = `<div class="program-item"><div class="info"><div class="name">${t('empty_no_tweaks')}</div><div class="desc">${t('hint_check_tweaks')}</div></div></div>`;
        updateTweakCounts();
        return;
    }

    const visibleTweaks = allTweaks.filter(t => (showRemovedTweaks || !t.removed) && !(hideActiveTweaks && t.stateKnown && t.active) && t.category !== 'debloat');

    const categories = {};
    visibleTweaks.forEach(t => {
        if (!categories[t.category]) categories[t.category] = [];
        categories[t.category].push(t);
    });

    const categoryNames = {
        power: t('cat_power'),
        privacy: t('cat_privacy'),
        visual: t('cat_visual'),
        services: t('cat_services'),
        edge: t('cat_edge'),
        gaming: t('cat_gaming'),
        security: t('cat_security'),
        debloat: t('cat_debloat')
    };

    let html = '';
    for (const [cat, tweaks] of Object.entries(categories)) {
        html += `<div style="padding:8px 16px;background:var(--bg-secondary);font-weight:600;color:var(--accent-primary);font-size:var(--fs-sm);">${categoryNames[cat] || cat}</div>`;
        tweaks.forEach((t) => {
            const idx = allTweaks.indexOf(t);
            html += tweakRowHtml(t, idx);
        });
    }
    list.innerHTML = html;
    updateTweakCounts();
}

// Shared row markup for a single tweak. Used by both renderTweaks() and renderDebloat().
// `opts.stateLabels` lets callers override the inactive/active labels (e.g. debloat uses
// "Zainstalowany"/"Usunięty" instead of "Nieaktywny"/"Aktywny").
function tweakRowHtml(tweak, idx, opts) {
    opts = opts || {};
    const labels = opts.stateLabels || {};
    const riskClass = tweak.risk === 'medium' ? 'risk-medium' : 'risk-low';
    const riskLabel = tweak.risk === 'medium' ? t('risk_medium') : t('risk_low');
    const situationalBadge = tweak.situational
        ? `<span class="badge badge-warn" title="${t('badge_situational')}">${t('badge_situational')}</span>`
        : '';

    // Determine live system state and preset affiliation.
    const inPreset = !!tweak.selected;
    const isLive = tweak.stateKnown && tweak.active;
    const isInactive = tweak.stateKnown && !tweak.active;

    let borderColor = 'var(--border-color)';
    let liveBadge = '';
    let presetBadge = '';
    let opacity = '1';

    if (tweak.removed) {
        borderColor = 'var(--text-muted)';
        opacity = '0.5';
        liveBadge = `<span class="status-pill">${t('tweak_removed_system')}</span>`;
    } else if (isLive) {
        borderColor = 'var(--success)';
        liveBadge = `<span class="status-pill is-success">${labels.active || t('tweak_active')}</span>`;
    } else if (isInactive) {
        borderColor = 'var(--warn)';
        liveBadge = `<span class="status-pill is-warn">${labels.inactive || t('tweak_inactive')}</span>`;
    } else {
        liveBadge = `<span class="status-pill">${t('tweak_unchecked')}</span>`;
    }

    if (inPreset) {
        presetBadge = `<span class="badge badge-info" title="${t('hint_preset_member')}">☑ ${t('badge_in_preset')}</span>`;
    }

    const stateMessage = tweak.stateMsg ? tMsg(tweak.stateMsg.code, tweak.stateMsg.params, tweak.stateMsg.message) : '';
    const stateSuffix = stateMessage && stateMessage.toLowerCase() !== String(liveBadge).toLowerCase()
        ? ' — ' + stateMessage
        : '';

    const revertBtn = tweak.revertible
        ? `<button class="btn-secondary btn-sm" onclick="revertTweak(${idx})" style="margin-left:8px;" title="${t('cfg_revert_title')}">${t('btn_revert')}</button>`
        : '';

    const badgesRow = `<div style="display:flex;gap:6px;flex-wrap:wrap;margin-top:4px;">${liveBadge}${presetBadge}${stateSuffix ? `<span style="color:var(--text-muted);font-size:var(--fs-xs);">${stateSuffix}</span>` : ''}</div>`;

    return `
        <div class="program-item" style="border-left:3px solid ${borderColor};opacity:${opacity};">
            <input type="checkbox" ${tweak.selected ? 'checked' : ''} onchange="toggleTweak(${idx}, this.checked)" ${tweak.removed ? 'disabled' : ''} title="${t('hint_checkbox_preset')}">
            <div class="info">
                <div class="name">${tweakText(tweak, 'name')}</div>
                <div class="desc">${tweakText(tweak, 'description')}</div>
                ${badgesRow}
            </div>
            ${situationalBadge}
            <span class="risk-badge ${riskClass}">${riskLabel}</span>
            ${revertBtn}
        </div>
    `;
}

// --- Debloat ---
function renderDebloat() {
    const list = document.getElementById('debloat-list');
    if (!list) return;

    const debloatTweaks = allTweaks.filter(t => t.category === 'debloat');

    if (!debloatTweaks.length) {
        list.innerHTML = `<div class="program-item"><div class="info"><div class="name">${t('empty_no_debloat')}</div><div class="desc">${t('hint_check_debloat')}</div></div></div>`;
        updateDebloatCounts();
        return;
    }

    // Debloat rows show a single clean status ("Usunięty"/"Zainstalowany").
    const labels = { active: t('tweak_removed'), inactive: t('tweak_installed') };

    // Group order; anything with no/unknown group falls into "Inne".
    const groupOrder = ['Xbox', 'Multimedia', 'Bing', 'Komunikacja', 'Narzędzia i inne'];
    const groups = {};
    groupOrder.forEach(g => groups[g] = []);
    const otherGroup = [];

    debloatTweaks.forEach(t => {
        // When hiding removed apps, skip ones already removed/active.
        if (hideRemovedDebloat && t.stateKnown && t.active) return;
        if (t.group && groups[t.group]) {
            groups[t.group].push(t);
        } else {
            otherGroup.push(t);
        }
    });

    let html = '';
    const renderGroup = (heading, tweaks) => {
        if (!tweaks.length) return;
        html += `<div style="padding:8px 16px;background:var(--bg-secondary);font-weight:600;color:var(--accent-primary);font-size:var(--fs-sm);">${heading}</div>`;
        tweaks.forEach(t => {
            const idx = allTweaks.indexOf(t);
            html += tweakRowHtml(t, idx, { stateLabels: labels });
        });
    };

    // Group ids come from tweaks.json; brand names (Xbox, Bing) stay as they are.
    const groupLabels = {
        Multimedia: t('cfg_debloat_group_multimedia'),
        Komunikacja: t('cfg_debloat_group_communication'),
        'Narzędzia i inne': t('cfg_debloat_group_tools_other')
    };
    groupOrder.forEach(g => renderGroup(groupLabels[g] || g, groups[g]));
    renderGroup(t('cat_other'), otherGroup);

    list.innerHTML = html;
    updateDebloatCounts();
}

function toggleRemovedDebloat() {
    hideRemovedDebloat = !hideRemovedDebloat;
    updateRemovedDebloatButton();
    renderDebloat();
}

function updateRemovedDebloatButton() {
    const btn = document.getElementById('toggle-removed-debloat');
    if (btn) btn.textContent = '👁 ' + (hideRemovedDebloat ? t('btn_show_removed_deb') : t('btn_hide_removed_deb'));
}

function updateDebloatCounts() {
    const debloatTweaks = allTweaks.filter(t => t.category === 'debloat');
    const selected = debloatTweaks.filter(t => t.selected).length;
    const active = debloatTweaks.filter(t => t.stateKnown && t.active).length;
    const toApply = debloatTweaks.filter(t => t.selected && !(t.stateKnown && t.active)).length;
    const countEl = document.getElementById('debloat-count');
    const activeEl = document.getElementById('debloat-active-count');
    const applyEl = document.getElementById('debloat-apply-count');
    if (countEl) countEl.textContent = tn('count_cfg_selected', selected);
    if (activeEl) activeEl.textContent = tn('count_cfg_already_removed', active);
    if (applyEl) applyEl.textContent = tn('count_cfg_debloat_to_remove', toApply);
}

// --- Script review --------------------------------------------------------
// Tweaks come from data\tweaks.json, which the user imports or edits, and run
// elevated. The backend refuses to run a tweak whose command was not approved
// in its current form, so new or changed tweaks are shown here first.
// Resolves to { ids, approve: [hashes] } or null when cancelled.
async function reviewTweakScripts(list, plainConfirmMessage) {
    const ids = list.map(tw => tw.id);
    const preview = await window.api.preview_tweaks(JSON.stringify(ids));
    if (!preview || preview.error) {
        uiToast(tErr(preview) || t('cfg_err_preview_tweaks'), { kind: 'error' });
        return null;
    }
    const pending = (preview.tweaks || []).filter(p => !p.approved);
    if (pending.length === 0) {
        return (await uiConfirm(plainConfirmMessage, { confirmLabel: t('run_program') })) ? { ids, approve: [] } : null;
    }

    const overlay = document.getElementById('tweak-review-overlay');
    const listEl = document.getElementById('tweak-review-list');
    const check = document.getElementById('tweak-review-check');
    const runBtn = document.getElementById('tweak-review-run');
    const cancelBtn = document.getElementById('tweak-review-cancel');

    listEl.innerHTML = pending.map(p => {
        const risk = String(p.risk || '').toLowerCase();
        return `<section class="tweak-review-item">
            <header><span>${escapeHtml(p.name || p.id)}</span>
                <span class="tag">${escapeHtml(p.type || '')}</span>
                ${risk ? `<span class="tag risk-${escapeHtml(risk)}">${escapeHtml(risk)}</span>` : ''}
            </header>
            <pre>${escapeHtml(p.command || '')}</pre>
        </section>`;
    }).join('');
    check.checked = false;
    runBtn.disabled = true;
    overlay.classList.add('active');
    check.focus();

    return new Promise(resolve => {
        const finish = (result) => {
            overlay.classList.remove('active');
            check.removeEventListener('change', onCheck);
            runBtn.removeEventListener('click', onRun);
            cancelBtn.removeEventListener('click', onCancel);
            document.removeEventListener('keydown', onKey, true);
            resolve(result);
        };
        const onCheck = () => { runBtn.disabled = !check.checked; };
        const onRun = () => { if (check.checked) finish({ ids, approve: pending.map(p => p.hash) }); };
        const onCancel = () => finish(null);
        const onKey = (e) => { if (e.key === 'Escape') { e.preventDefault(); finish(null); } };
        check.addEventListener('change', onCheck);
        runBtn.addEventListener('click', onRun);
        cancelBtn.addEventListener('click', onCancel);
        document.addEventListener('keydown', onKey, true);
    });
}

async function applyDebloat() {
    const toApply = allTweaks.filter(t => t.category === 'debloat' && t.selected && !(t.stateKnown && t.active));
    if (toApply.length === 0) {
        uiToast(t('val_no_debloat'), { kind: 'warn' });
        return;
    }

    const request = await reviewTweakScripts(toApply, tn('count_cfg_confirm_debloat', toApply.length) + '\n\n' + t('confirm_debloat_warn'));
    if (!request) return;

    const overlay = document.getElementById('progress-overlay');
    const bar = document.getElementById('progress-bar');
    const status = document.getElementById('progress-status');
    const log = document.getElementById('progress-log');
    const closeBtn = document.getElementById('progress-close');

    overlay.classList.add('active');
    log.innerHTML = '';
    closeBtn.style.display = 'none';
    bar.style.width = '0%';
    status.textContent = t('progress_remove_apps');

    try {
        const startResult = await window.api.apply_tweaks(JSON.stringify(request));

        if (startResult.error) {
            log.innerHTML = `<div class="err">❌ ${escapeHtml(tErr(startResult))}</div>`;
            closeBtn.style.display = 'block';
            return;
        }

        pollLiveProgress('check_tweak_progress', startResult.sessionId, t('progress_remove_apps'), async () => {
            tweakStatesScanned = false;
            await scanTweakStates(true);
            renderDebloat();
        });

    } catch (e) {
        log.innerHTML = `<div class="err">❌ ${String(e)}</div>`;
        closeBtn.style.display = 'block';
    }
}

async function scanTweakStates(force) {
    if (tweakStatesScanning) return;
    if (tweakStatesScanned && !force) return;

    const status = document.getElementById('tweaks-scan-status');
    tweakStatesScanning = true;
    if (status) status.textContent = t('scan_checking');

    try {
        let states;
        if (typeof window.api.start_tweak_state_scan === 'function' &&
            typeof window.api.check_tweak_state_scan === 'function') {
            const started = await window.api.start_tweak_state_scan(!!force);
            if (started.error || !started.sessionId) {
                throw new Error(tErr(started) || t('cfg_err_scan_start'));
            }

            for (let attempt = 0; attempt < 1200; attempt++) {
                await new Promise(resolve => setTimeout(resolve, 100));
                const progress = await window.api.check_tweak_state_scan(started.sessionId);
                if (progress.error) {
                    throw new Error(progress.error);
                }
                if (status && progress.total) {
                    status.textContent = t('cfg_scan_progress', { done: progress.completed || 0, total: progress.total });
                }
                if (progress.done) {
                    states = progress.states || [];
                    break;
                }
            }
            if (!states) {
                throw new Error(t('cfg_err_scan_timeout'));
            }
        } else {
            // Compatibility fallback for older backends and standalone UI tests.
            states = await window.api.get_tweak_states();
            if (states.error) {
                throw new Error(states.error);
            }
        }

        const byId = {};
        (states || []).forEach(s => byId[s.id] = s);

        allTweaks.forEach(t => {
            const state = byId[t.id];
            if (!state) return;
            t.active = !!state.active;
            t.stateKnown = !!state.known;
            t.removed = !!state.removed;
            t.revertible = !!state.revertible;
            // Backend message code + Polish fallback, translated at render time.
            t.stateMsg = { code: state.code || '', params: state.params || null, message: state.message || '' };
            // Checkbox oznacza WYBÓR do zastosowania / część profilu,
            // a nie aktywność w systemie. Skan nie powinien sam zaznaczać
            // aktywnych tweaków, bo to myliło stan z wyborem użytkownika.
            t.selected = !t.removed && !!t.selected;
        });

        tweakStatesScanned = true;
        const activeCount = allTweaks.filter(t => t.stateKnown && t.active).length;
        const removedCount = allTweaks.filter(t => t.removed).length;
        if (status) status.textContent = t('cfg_scan_summary', { active: activeCount, total: allTweaks.length, removed: removedCount });
        renderTweaks();
        renderDebloat();
        updateRemovedButton();
    } catch (e) {
        console.error('Failed to scan tweak states:', e);
        if (status) status.textContent = t('cfg_scan_error', { error: String(e) });
    } finally {
        tweakStatesScanning = false;
    }
}

function toggleTweak(idx, checked) {
    allTweaks[idx].selected = !!checked;
    updateTweakCounts();
}

function updateTweakCounts() {
    const visibleTweaks = allTweaks.filter(t => showRemovedTweaks || !t.removed);
    const selected = visibleTweaks.filter(t => t.selected).length;
    const active = allTweaks.filter(t => t.stateKnown && t.active).length;
    const toApply = visibleTweaks.filter(t => t.selected && !(t.stateKnown && t.active) && !t.removed).length;
    document.getElementById('tweaks-count').textContent = tn('count_cfg_selected', selected);
    document.getElementById('tweaks-active-count').textContent = tn('count_cfg_already_active', active);
    document.getElementById('tweaks-apply-count').textContent = tn('count_cfg_tweaks_to_apply', toApply);
}

function toggleRemoved() {
    showRemovedTweaks = !showRemovedTweaks;
    updateRemovedButton();
    renderTweaks();
}

function toggleActive() {
    hideActiveTweaks = !hideActiveTweaks;
    updateActiveButton();
    renderTweaks();
}

function updateActiveButton() {
    const btn = document.getElementById('toggle-active-btn');
    if (btn) btn.textContent = '👁 ' + (hideActiveTweaks ? t('btn_show_active_tweaks') : t('btn_hide_active_tweaks'));
}

function updateRemovedButton() {
    const btn = document.getElementById('toggle-removed-btn');
    const removedCount = allTweaks.filter(t => t.removed).length;
    if (removedCount > 0) {
        btn.style.display = 'inline-block';
        btn.textContent = '👁 ' + t(showRemovedTweaks ? 'cfg_btn_hide_cut_count' : 'cfg_btn_show_cut_count', { count: removedCount });
    } else {
        btn.style.display = 'none';
    }
}

async function applyTweaks() {
    const selected = allTweaks.filter(t => t.selected);
    if (selected.length === 0) {
        uiToast(t('val_no_tweaks'), { kind: 'warn' });
        return;
    }

    const toApply = selected.filter(t => !(t.stateKnown && t.active));
    const alreadyActive = selected.filter(t => t.stateKnown && t.active);

    if (toApply.length === 0) {
        uiToast(tn('count_cfg_all_tweaks_active', alreadyActive.length), { kind: 'info' });
        return;
    }

    const question = tn('count_cfg_confirm_apply', toApply.length);
    const msg = alreadyActive.length > 0
        ? `${question}\n${tn('count_cfg_already_active_skipped', alreadyActive.length)}\n\n${t('confirm_apply_restart')}`
        : `${question}\n\n${t('confirm_apply_restart')}`;

    const request = await reviewTweakScripts(toApply, msg);
    if (!request) return;

    const overlay = document.getElementById('progress-overlay');
    const bar = document.getElementById('progress-bar');
    const status = document.getElementById('progress-status');
    const log = document.getElementById('progress-log');
    const closeBtn = document.getElementById('progress-close');

    overlay.classList.add('active');
    log.innerHTML = '';
    closeBtn.style.display = 'none';
    bar.style.width = '0%';
    status.textContent = t('progress_apply_tweaks');

    try {
        const startResult = await window.api.apply_tweaks(JSON.stringify(request));

        if (startResult.error) {
            log.innerHTML = `<div class="err">❌ ${escapeHtml(tErr(startResult))}</div>`;
            closeBtn.style.display = 'block';
            return;
        }

        pollLiveProgress('check_tweak_progress', startResult.sessionId, t('progress_apply_tweaks'), async () => {
            tweakStatesScanned = false;
            await scanTweakStates(true);
        });

    } catch (e) {
        log.innerHTML = `<div class="err">❌ ${String(e)}</div>`;
        closeBtn.style.display = 'block';
    }
}

async function revertTweak(idx) {
    const tweak = allTweaks[idx];
    if (!tweak) return;

    if (!(await uiConfirm(t('cfg_confirm_revert', { name: tweakText(tweak, 'name') }) + '\n\n' + t('cfg_revert_backup_note'), { confirmLabel: t('cfg_btn_revert') }))) {
        return;
    }

    const overlay = document.getElementById('progress-overlay');
    const bar = document.getElementById('progress-bar');
    const status = document.getElementById('progress-status');
    const log = document.getElementById('progress-log');
    const closeBtn = document.getElementById('progress-close');

    overlay.classList.add('active');
    log.innerHTML = '';
    closeBtn.style.display = 'none';
    bar.style.width = '0%';
    status.textContent = t('progress_revert');

    try {
        const result = JSON.parse(await window.api.revert_tweak(tweak.id));
        const cls = result.success ? 'ok' : 'err';
        const icon = result.success ? '✅' : '❌';
        log.innerHTML += `<div class="${cls}">${icon} ${escapeHtml(tweakText(tweak, 'name'))} — ${escapeHtml(tMsg(result.code, result.params, result.message))}</div>`;
        bar.style.width = '100%';
        status.textContent = result.success ? t('revert_ok') : t('revert_fail');

        if (result.success) {
            tweakStatesScanned = false;
            await scanTweakStates(true);
        }

        closeBtn.style.display = 'block';
    } catch (e) {
        log.innerHTML = `<div class="err">${escapeHtml(t('mng_err_generic', { error: String(e) }))}</div>`;
        closeBtn.style.display = 'block';
    }
}
