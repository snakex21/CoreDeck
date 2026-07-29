// --- Install ---
async function startInstall() {
    const selected = allPrograms.filter(p => p.selected).map(p => {
        if (p.variants && p.variants.length > 0) {
            let vi = (typeof p.selectedVariant === 'number' && p.selectedVariant >= 0 && p.selectedVariant < p.variants.length)
                ? p.selectedVariant
                : ((typeof p.defaultVariant === 'number' && p.defaultVariant >= 0 && p.defaultVariant < p.variants.length) ? p.defaultVariant : 0);
            const v = p.variants[vi];
            return Object.assign({}, p, {
                selectedVariant: vi,
                folderPath: v.folderPath || p.folderPath,
                installType: v.installType || p.installType
            });
        }
        return p;
    });
    if (selected.length === 0) {
        alert(t('val_no_programs'));
        return;
    }

    const targetDir = document.getElementById('target-dir').value || 'C:\\Tools';

    if (!confirm(`${t('confirm_install_start')} ${selected.length} ${t('confirm_install_mid')}?\n\nPortable/tool/addon: kopiowanie do:\n${targetDir}\nInstalator: uruchomienie instalatora i czekanie na zamknięcie\nSkrypt: uruchomienie skryptu\n\nKontynuować?`)) {
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
    status.textContent = t('cfg_progress_preparing');

    try {
        // Start async installation — returns sessionId immediately
        const startResult = await window.api.install_selected_programs(
            JSON.stringify(selected),
            targetDir
        );

        if (startResult.error) {
            log.innerHTML = `<div class="err">❌ ${startResult.error}</div>`;
            closeBtn.style.display = 'block';
            return;
        }

        const sessionId = startResult.sessionId;
        if (!sessionId) {
            log.innerHTML = `<div class="err">❌ No session ID</div>`;
            closeBtn.style.display = 'block';
            return;
        }

        pollLiveProgress('check_install_progress', sessionId, t('cfg_progress_title'));

    } catch (e) {
        log.innerHTML = `<div class="err">❌ ${String(e)}</div>`;
        closeBtn.style.display = 'block';
    }
}

// Generic live-progress polling helper for async operations (install, tweaks, debloat).
let _currentPollSession = null; // { sessionId, timeoutId }

function cancelCurrentOperation() {
    if (_currentPollSession) {
        window.api.cancel_operation(_currentPollSession.sessionId);
        const cancelBtn = document.getElementById('progress-cancel');
        if (cancelBtn) cancelBtn.style.display = 'none';
    }
}

function closeProgress() {
    if (_currentPollSession && _currentPollSession.timeoutId) {
        clearTimeout(_currentPollSession.timeoutId);
    }
    _currentPollSession = null;
    document.getElementById('progress-overlay').classList.remove('active');
    const cancelBtn = document.getElementById('progress-cancel');
    if (cancelBtn) cancelBtn.style.display = 'none';
}

// checkFnName is the window.api method name (e.g. 'check_install_progress' or 'check_tweak_progress').
function pollLiveProgress(checkFnName, sessionId, title, onComplete) {
    const bar = document.getElementById('progress-bar');
    const status = document.getElementById('progress-status');
    const log = document.getElementById('progress-log');
    const closeBtn = document.getElementById('progress-close');
    const cancelBtn = document.getElementById('progress-cancel');
    const titleEl = document.querySelector('#progress-overlay h3');
    if (titleEl && title) titleEl.textContent = title;
    if (cancelBtn) cancelBtn.style.display = 'inline-block';
    let displayedCount = 0;

    const poll = async () => {
        try {
            const s = await window.api[checkFnName](sessionId);

            if (s.error && !s.done) {
                log.innerHTML += `<div class="err">❌ ${s.error}</div>`;
            }

            if (s.total > 0) {
                const pct = Math.round((s.current / s.total) * 100);
                bar.style.width = pct + '%';
            }

            if (s.program && !s.done) {
                status.textContent = `⏳ ${s.current}/${s.total} — ${s.program}...`;
            }

            // Show new results
            if (s.results && s.results.length > displayedCount) {
                s.results.slice(displayedCount).forEach(r => {
                    log.innerHTML += renderProgressResult(r);
                    displayedCount++;
                });
                log.scrollTop = log.scrollHeight;
            }

            if (s.done) {
                // Final batch
                if (s.results && s.results.length > displayedCount) {
                    s.results.slice(displayedCount).forEach(r => {
                        log.innerHTML += renderProgressResult(r);
                        displayedCount++;
                    });
                    log.scrollTop = log.scrollHeight;
                }
                bar.style.width = '100%';
                if (cancelBtn) cancelBtn.style.display = 'none';
                const ok = s.results ? s.results.filter(r => r.success).length : 0;
                const total = s.results ? s.results.length : 0;
                if (s.cancelled) {
                    status.textContent = `${t('cfg_cancelled')} ${ok}/${total}`;
                } else {
                    status.textContent = `${t('progress_finished')} ${ok}/${total} ${t('progress_success')}`;
                }
                closeBtn.style.display = 'block';
                _currentPollSession = null;
                if (onComplete) onComplete(s);
                return;
            }

            _currentPollSession.timeoutId = setTimeout(poll, 500);
        } catch (e) {
            log.innerHTML += `<div class="err">❌ ${String(e)}</div>`;
            closeBtn.style.display = 'block';
        }
    };

    _currentPollSession = { sessionId, timeoutId: null };
    _currentPollSession.timeoutId = setTimeout(poll, 500);
}

async function openFolder() {
    const targetDir = document.getElementById('target-dir').value || 'C:\\Tools';
    await window.api.open_target_folder(targetDir);
}

function loadTargetDir() {
    const saved = localStorage.getItem('configurator_target_dir');
    if (saved) {
        document.getElementById('target-dir').value = saved;
    } else {
        const drive = systemInfo ? (systemInfo.os || '').includes('C') ? 'C' : 'C' : 'C';
        document.getElementById('target-dir').value = drive + ':\\Tools';
    }
    document.getElementById('target-dir').addEventListener('change', (e) => {
        localStorage.setItem('configurator_target_dir', e.target.value);
    });
}
