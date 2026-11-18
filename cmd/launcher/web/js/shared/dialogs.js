// In-app dialogs and toasts.
//
// CoreDeck runs inside WebView2, where window.alert/confirm/prompt show the
// browser's own "Message from localhost:8000" box. These helpers replace them
// with dialogs styled like the rest of the app:
//
//   await uiAlert(message, { title, kind })            -> undefined
//   await uiConfirm(message, { title, confirmLabel, danger }) -> true/false
//   await uiPrompt(message, defaultValue, { title })   -> string or null
//   await uiChoose(message, [{ label, hint, value }], { title, extra }) -> value or null
//   uiToast(message, { kind: 'success'|'error'|'info'|'warn'|'neutral', timeout })
//
// Keyboard: Enter confirms, Escape cancels, Tab stays inside the dialog and
// focus returns to the element that opened it. Dialogs opened while another
// one is visible wait their turn.
(function () {
    'use strict';

    function tr(key, fallback) {
        try {
            if (typeof window.t === 'function') {
                const value = window.t(key);
                if (value && value !== key) return value;
            }
        } catch (_) { /* translations not loaded */ }
        return fallback;
    }

    const FOCUSABLE = 'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
    let queue = Promise.resolve();
    let idCounter = 0;

    function el(tag, className, text) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (text !== undefined && text !== null) node.textContent = String(text);
        return node;
    }

    // open() builds one dialog and resolves with whatever finish() receives.
    // spec: { kind, title, message, input, choices, buttons: [{label, value, variant, autofocus}], cancelValue, extra }
    function open(spec) {
        const run = () => new Promise(resolve => {
            const opener = document.activeElement;
            const id = `cd-dialog-${++idCounter}`;

            const backdrop = el('div', 'cd-dialog-backdrop');
            const dialog = el('div', `cd-dialog cd-dialog--${spec.kind || 'info'}`);
            dialog.setAttribute('role', spec.role || 'dialog');
            dialog.setAttribute('aria-modal', 'true');
            dialog.tabIndex = -1;

            if (spec.title) {
                const title = el('h2', 'cd-dialog-title', spec.title);
                title.id = `${id}-title`;
                dialog.setAttribute('aria-labelledby', title.id);
                dialog.appendChild(title);
            }
            if (spec.message) {
                const message = el('p', 'cd-dialog-message', spec.message);
                message.id = `${id}-message`;
                dialog.setAttribute('aria-describedby', message.id);
                dialog.appendChild(message);
            }

            if (spec.body instanceof Node) {
                const body = el('div', 'cd-dialog-body');
                body.appendChild(spec.body);
                dialog.appendChild(body);
            }
            if (spec.wide) dialog.classList.add('cd-dialog--wide');

            let input = null;
            if (spec.input) {
                input = el('input', 'cd-dialog-input');
                input.type = 'text';
                input.value = spec.input.value || '';
                input.setAttribute('aria-label', spec.title || spec.message || '');
                dialog.appendChild(input);
            }

            if (Array.isArray(spec.choices) && spec.choices.length) {
                const list = el('div', 'cd-dialog-choices');
                list.setAttribute('role', 'list');
                spec.choices.forEach((choice, index) => {
                    const button = el('button', 'cd-dialog-choice');
                    button.type = 'button';
                    button.setAttribute('role', 'listitem');
                    button.appendChild(el('span', 'cd-dialog-choice-label', choice.label));
                    if (choice.hint) button.appendChild(el('span', 'cd-dialog-choice-hint', choice.hint));
                    button.addEventListener('click', () => finish(choice.value));
                    if (index === 0) button.dataset.autofocus = '1';
                    list.appendChild(button);
                });
                dialog.appendChild(list);
            }

            const actions = el('div', 'cd-dialog-actions');
            let primary = null;
            (spec.buttons || []).forEach(def => {
                const button = el('button', `cd-dialog-btn ${def.variant === 'primary' ? 'btn-primary' : def.variant === 'danger' ? 'btn-primary cd-dialog-btn--danger' : 'btn-secondary'}`, def.label);
                button.type = 'button';
                button.dataset.role = def.role || '';
                button.addEventListener('click', () => {
                    if (typeof def.onClick === 'function') def.onClick();
                    finish(def.role === 'confirm' && input ? input.value : def.value);
                });
                if (def.role === 'confirm') primary = button;
                actions.appendChild(button);
            });
            dialog.appendChild(actions);
            backdrop.appendChild(dialog);

            let settled = false;
            function finish(value) {
                if (settled) return;
                settled = true;
                document.removeEventListener('keydown', onKey, true);
                backdrop.classList.add('is-closing');
                const remove = () => backdrop.remove();
                backdrop.addEventListener('animationend', remove, { once: true });
                setTimeout(remove, 220);
                if (opener && typeof opener.focus === 'function' && opener.isConnected) {
                    try { opener.focus({ preventScroll: true }); } catch (_) { /* ignore */ }
                }
                resolve(value);
            }

            function onKey(event) {
                if (event.key === 'Escape') {
                    event.preventDefault();
                    event.stopPropagation();
                    finish(spec.cancelValue);
                    return;
                }
                if (event.key === 'Enter' && !event.shiftKey && !event.isComposing) {
                    const target = event.target;
                    // Enter on a focused button activates that button
                    // (e.g. Cancel); anywhere else it confirms.
                    if (target && target.tagName === 'BUTTON' && dialog.contains(target)) {
                        event.preventDefault();
                        event.stopPropagation();
                        target.click();
                        return;
                    }
                    if (primary) {
                        event.preventDefault();
                        event.stopPropagation();
                        primary.click();
                    }
                    return;
                }
                if (event.key === 'Tab') {
                    const items = Array.from(dialog.querySelectorAll(FOCUSABLE));
                    if (!items.length) { event.preventDefault(); return; }
                    const first = items[0];
                    const last = items[items.length - 1];
                    if (event.shiftKey && (document.activeElement === first || !dialog.contains(document.activeElement))) {
                        event.preventDefault();
                        last.focus();
                    } else if (!event.shiftKey && (document.activeElement === last || !dialog.contains(document.activeElement))) {
                        event.preventDefault();
                        first.focus();
                    }
                }
            }

            backdrop.addEventListener('mousedown', event => {
                if (event.target === backdrop) finish(spec.cancelValue);
            });
            document.addEventListener('keydown', onKey, true);
            document.body.appendChild(backdrop);

            const focusTarget = input || dialog.querySelector('[data-autofocus]') || primary || dialog.querySelector(FOCUSABLE) || dialog;
            // A macrotask (not requestAnimationFrame, which background
            // pages pause) so focus lands once the dialog is in the DOM.
            setTimeout(() => {
                focusTarget.focus();
                if (input) input.select();
            }, 0);
        });

        const result = queue.then(run, run);
        queue = result.then(() => undefined, () => undefined);
        return result;
    }

    function uiAlert(message, options) {
        const opts = options || {};
        const kind = opts.kind || 'info';
        return open({
            kind,
            role: kind === 'error' ? 'alertdialog' : 'dialog',
            title: opts.title || (kind === 'error' ? tr('dlg_error_title', 'Coś poszło nie tak') : tr('dlg_info_title', 'Informacja')),
            message,
            buttons: [{ label: opts.okLabel || tr('dlg_ok', 'OK'), role: 'confirm', variant: 'primary', value: undefined }],
            cancelValue: undefined
        });
    }

    function uiConfirm(message, options) {
        const opts = options || {};
        return open({
            kind: opts.danger ? 'danger' : 'question',
            role: 'alertdialog',
            title: opts.title || tr('dlg_confirm_title', 'Potwierdź'),
            message,
            buttons: [
                { label: opts.cancelLabel || tr('dlg_cancel', 'Anuluj'), role: 'cancel', variant: 'secondary', value: false },
                { label: opts.confirmLabel || (opts.danger ? tr('dlg_delete', 'Usuń') : tr('dlg_ok', 'OK')), role: 'confirm', variant: opts.danger ? 'danger' : 'primary', value: true }
            ],
            cancelValue: false
        }).then(Boolean);
    }

    function uiPrompt(message, defaultValue, options) {
        const opts = options || {};
        return open({
            kind: 'question',
            title: opts.title || message,
            message: opts.title ? message : '',
            input: { value: defaultValue == null ? '' : String(defaultValue) },
            buttons: [
                { label: opts.cancelLabel || tr('dlg_cancel', 'Anuluj'), role: 'cancel', variant: 'secondary', value: null },
                { label: opts.confirmLabel || tr('dlg_save', 'Zapisz'), role: 'confirm', variant: 'primary' }
            ],
            cancelValue: null
        });
    }

    // Generic dialog: uiDialog({ title, message, body: Node, kind, wide,
    //   buttons: [{ label, value, variant: 'primary'|'secondary'|'danger', confirm: true }],
    //   cancelValue }) -> value of the clicked button (cancelValue on Esc).
    function uiDialog(options) {
        const opts = options || {};
        return open({
            kind: opts.kind || 'info',
            role: opts.role || 'dialog',
            title: opts.title,
            message: opts.message,
            body: opts.body,
            wide: !!opts.wide,
            buttons: (opts.buttons || []).map(b => ({
                label: b.label, value: b.value, variant: b.variant || 'secondary', role: b.confirm ? 'confirm' : (b.role || 'button')
            })),
            cancelValue: opts.cancelValue
        });
    }

    function uiChoose(message, choices, options) {
        const opts = options || {};
        const buttons = [{ label: opts.cancelLabel || tr('dlg_cancel', 'Anuluj'), role: 'cancel', variant: 'secondary', value: null }];
        if (opts.extra) buttons.unshift({ label: opts.extra.label, role: 'extra', variant: 'secondary', value: opts.extra.value });
        return open({
            kind: 'question',
            title: opts.title || tr('dlg_choose_title', 'Wybierz'),
            message,
            choices,
            buttons,
            cancelValue: null
        });
    }

    // ---- Toasts ----
    let region = null;
    function toastRegion() {
        if (region && region.isConnected) return region;
        region = el('div', 'cd-toast-region');
        region.setAttribute('aria-live', 'polite');
        region.setAttribute('aria-relevant', 'additions');
        document.body.appendChild(region);
        return region;
    }

    const TOAST_ICONS = { success: '✓', error: '!', warn: '!', info: 'i', neutral: '–' };

    function uiToast(message, options) {
        const opts = typeof options === 'string' ? { kind: options } : (options || {});
        const kind = TOAST_ICONS[opts.kind] ? opts.kind : 'info';
        const timeout = Number.isFinite(opts.timeout) ? opts.timeout : (kind === 'error' ? 8000 : 4000);

        const toast = el('div', `cd-toast cd-toast--${kind}`);
        toast.setAttribute('role', kind === 'error' ? 'alert' : 'status');
        toast.appendChild(el('span', 'cd-toast-icon', TOAST_ICONS[kind])).setAttribute('aria-hidden', 'true');
        const body = el('div', 'cd-toast-body');
        if (opts.title) body.appendChild(el('strong', 'cd-toast-title', opts.title));
        body.appendChild(el('span', 'cd-toast-text', message));
        toast.appendChild(body);
        const close = el('button', 'cd-toast-close', '×');
        close.type = 'button';
        close.setAttribute('aria-label', tr('dlg_close', 'Zamknij'));
        toast.appendChild(close);

        let timer = null;
        const dismiss = () => {
            clearTimeout(timer);
            if (!toast.isConnected) return;
            toast.classList.add('is-leaving');
            setTimeout(() => toast.remove(), 200);
        };
        const arm = () => { if (timeout > 0) timer = setTimeout(dismiss, timeout); };
        close.addEventListener('click', dismiss);
        toast.addEventListener('mouseenter', () => clearTimeout(timer));
        toast.addEventListener('mouseleave', arm);
        toastRegion().appendChild(toast);
        arm();
        const update = (text, kind) => {
            toast.querySelector('.cd-toast-text').textContent = String(text);
            if (kind && TOAST_ICONS[kind]) {
                toast.className = `cd-toast cd-toast--${kind}`;
                toast.querySelector('.cd-toast-icon').textContent = TOAST_ICONS[kind];
            }
        };
        return { dismiss, update, element: toast };
    }

    window.CoreDialog = { alert: uiAlert, confirm: uiConfirm, prompt: uiPrompt, choose: uiChoose, dialog: uiDialog, toast: uiToast };
    window.uiDialog = uiDialog;
    window.uiAlert = uiAlert;
    window.uiConfirm = uiConfirm;
    window.uiPrompt = uiPrompt;
    window.uiChoose = uiChoose;
    window.uiToast = uiToast;
})();
