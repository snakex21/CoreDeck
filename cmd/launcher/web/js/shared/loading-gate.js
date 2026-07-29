const coreDeckLoadingGate = (() => {
    const overlay = document.getElementById('coredeck-loading');
    const status = document.getElementById('coredeck-loading-status');
    if (!overlay) return { setStatus() {}, finish() {} };

    let finished = false;
    let visibleAt = 0;
    const showTimer = setTimeout(() => {
        if (finished) return;
        visibleAt = performance.now();
        overlay.classList.add('is-visible');
        overlay.setAttribute('aria-hidden', 'false');
    }, 140);

    function finish() {
        if (finished) return;
        finished = true;
        clearTimeout(showTimer);
        clearTimeout(hardTimer);

        if (!visibleAt) {
            overlay.remove();
            return;
        }

        const minimumVisibleTime = 360;
        const remaining = Math.max(0, minimumVisibleTime - (performance.now() - visibleAt));
        setTimeout(() => {
            overlay.classList.add('is-leaving');
            overlay.setAttribute('aria-hidden', 'true');
            setTimeout(() => overlay.remove(), 240);
        }, remaining);
    }

    const hardTimer = setTimeout(finish, 2500);
    return {
        setStatus(message) {
            if (status && message) status.textContent = message;
        },
        finish
    };
})();
