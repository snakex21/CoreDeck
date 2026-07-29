const coreDeckLoadingGate = (() => {
    const overlay = document.getElementById('coredeck-loading');
    const status = document.getElementById('coredeck-loading-status');
    if (!overlay) return { setStatus() {}, finish() {} };

    const brandMark = overlay.querySelector('.coredeck-loading-mark img');
    const brandMarkReady = !brandMark || (brandMark.complete && brandMark.naturalWidth > 0)
        ? Promise.resolve(true)
        : Promise.race([
            typeof brandMark.decode === 'function'
                ? brandMark.decode().then(() => true).catch(() => false)
                : new Promise(resolve => {
                    brandMark.addEventListener('load', () => resolve(true), { once: true });
                    brandMark.addEventListener('error', () => resolve(false), { once: true });
                }),
            new Promise(resolve => setTimeout(() => resolve(false), 500))
        ]);

    let finished = false;
    let visibleAt = 0;
    const showTimer = setTimeout(async () => {
        const markIsReady = await brandMarkReady;
        if (finished) return;
        // Never reveal a half-built loading screen. If the brand mark is late,
        // the normal app view is preferable to a logo popping in afterwards.
        if (!markIsReady) return;
        visibleAt = performance.now();
        overlay.classList.add('is-visible');
        overlay.setAttribute('aria-hidden', 'false');
    }, 550);

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
