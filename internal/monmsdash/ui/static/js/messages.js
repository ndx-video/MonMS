/**
 * MonMS Console — Global Message Strip
 *
 * Persistent status slot in the top navbar plus click-to-open history modal
 * (max 100 items, session-scoped). Adapted from Sentinel dashboard messages.js.
 *
 * API:
 *   monms.msg(text, type?, meta?)   — add message (type: info|success|error|warn)
 *   monms.msgOk(text, meta?)        — shorthand success
 *   monms.msgErr(text, meta?)       — shorthand error
 *   monms.msgWarn(text, meta?)      — shorthand warn
 *   monms.msgInfo(text, meta?)      — shorthand info
 *   monms.msgClear()                — clear history
 */
(function () {
    'use strict';

    var MAX_HISTORY = 100;

    var history = [];
    var stripEl = null;
    var modalEl = null;
    var historyListEl = null;

    var TYPE_CFG = {
        info:    { icon: 'info',         bg: 'bg-monms-surface-2',  border: 'border-monms-border',    text: 'text-monms-muted',   label: 'INFO' },
        success: { icon: 'check_circle', bg: 'bg-emerald-900/60',  border: 'border-emerald-700/40',  text: 'text-emerald-300',   label: 'OK' },
        error:   { icon: 'error',        bg: 'bg-red-900/60',      border: 'border-red-700/40',      text: 'text-red-300',       label: 'ERROR' },
        warn:    { icon: 'warning',      bg: 'bg-yellow-900/60',   border: 'border-yellow-700/40',   text: 'text-yellow-300',    label: 'WARN' }
    };

    function esc(s) {
        var d = document.createElement('div');
        d.textContent = s;
        return d.innerHTML;
    }

    function ts() {
        var d = new Date();
        return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    }

    function normalizeMeta(meta) {
        if (!meta || typeof meta !== 'object') return {};
        return {
            detail: meta.detail == null ? '' : String(meta.detail),
            trace: meta.trace == null ? null : meta.trace,
            copyText: meta.copyText == null ? '' : String(meta.copyText)
        };
    }

    function formatTrace(trace) {
        if (trace == null || trace === '') return '';
        if (typeof trace === 'string') return trace;
        try {
            return JSON.stringify(trace, null, 2);
        } catch (_e) {
            return String(trace);
        }
    }

    function buildCopyText(entry) {
        if (entry.copyText) return entry.copyText;
        var parts = ['[' + entry.type.toUpperCase() + '] ' + entry.time, entry.text];
        if (entry.detail) {
            parts.push('', 'Detail:', entry.detail);
        }
        var traceText = formatTrace(entry.trace);
        if (traceText) {
            parts.push('', 'Trace:', traceText);
        }
        return parts.join('\n');
    }

    function copyText(text, onDone) {
        if (!text) return;
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function () {
                if (typeof onDone === 'function') onDone();
            });
        }
    }

    function ensureStrip() {
        if (stripEl) return;
        stripEl = document.getElementById('monms-msg-strip');
        if (stripEl) return;
        stripEl = document.createElement('div');
        stripEl.id = 'monms-msg-strip';
        stripEl.className = 'min-w-0';
        var topbar = document.getElementById('app-topbar');
        if (topbar && topbar.children.length >= 2) {
            topbar.insertBefore(stripEl, topbar.children[1]);
        } else {
            document.body.prepend(stripEl);
        }
    }

    function buildStripHTML(entry, cfg) {
        return '<span class="material-symbols-outlined text-[16px] shrink-0">' + cfg.icon + '</span>' +
            '<span class="font-mono text-[10px] opacity-60 shrink-0">' + esc(entry.time) + '</span>' +
            '<span class="font-medium truncate flex-1">' + esc(entry.text) + '</span>' +
            '<span class="text-[10px] opacity-40 shrink-0 hidden sm:inline">' + history.length + ' msg' + (history.length === 1 ? '' : 's') + '</span>' +
            '<span class="material-symbols-outlined text-[14px] opacity-40 shrink-0">expand_more</span>';
    }

    function stripClasses(cfg) {
        return 'flex items-center gap-2 px-3 py-1.5 text-xs cursor-pointer select-none border border-monms-border/60 rounded-none min-w-0 overflow-hidden '
            + cfg.bg + ' ' + cfg.border + ' ' + cfg.text;
    }

    function renderStrip(entry) {
        ensureStrip();
        var cfg = TYPE_CFG[entry.type] || TYPE_CFG.info;
        stripEl.className = stripClasses(cfg);
        stripEl.title = 'Click to view message history';
        stripEl.innerHTML = buildStripHTML(entry, cfg);
        stripEl.onclick = openModal;
    }

    function ensureModal() {
        if (modalEl) return;
        modalEl = document.createElement('div');
        modalEl.id = 'monms-msg-modal';
        modalEl.className = 'fixed inset-0 bg-black/70 backdrop-blur-sm z-[100] flex items-start justify-center pt-12';
        modalEl.style.display = 'none';

        modalEl.innerHTML =
            '<div class="bg-monms-surface border border-monms-border w-full max-w-2xl max-h-[70vh] flex flex-col shadow-2xl" onclick="event.stopPropagation()">' +
            '  <div class="flex items-center justify-between px-4 py-3 border-b border-monms-border">' +
            '    <div class="flex items-center gap-2">' +
            '      <span class="material-symbols-outlined text-monms-accent text-[18px]">notifications</span>' +
            '      <span class="text-sm font-semibold text-monms-text">Message History</span>' +
            '      <span id="monms-msg-count" class="text-[10px] text-monms-dim font-mono"></span>' +
            '    </div>' +
            '    <div class="flex items-center gap-2">' +
            '      <button id="monms-msg-clear-btn" class="text-[10px] text-monms-dim hover:text-monms-muted px-2 py-1 hover:bg-monms-border/30 transition-colors">Clear</button>' +
            '      <button id="monms-msg-close-btn" class="text-monms-dim hover:text-monms-text transition-colors">' +
            '        <span class="material-symbols-outlined text-[18px]">close</span>' +
            '      </button>' +
            '    </div>' +
            '  </div>' +
            '  <div id="monms-msg-list" class="flex-1 overflow-y-auto p-2 space-y-1 min-h-0"></div>' +
            '</div>';

        document.body.appendChild(modalEl);
        historyListEl = document.getElementById('monms-msg-list');

        document.getElementById('monms-msg-close-btn').onclick = closeModal;
        document.getElementById('monms-msg-clear-btn').onclick = function () {
            history = [];
            renderHistory();
            closeModal();
            renderEmptyStrip();
        };

        modalEl.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') closeModal();
        });
        modalEl.onclick = closeModal;
    }

    function renderHistory() {
        if (!historyListEl) return;
        document.getElementById('monms-msg-count').textContent = history.length + ' / ' + MAX_HISTORY;

        if (history.length === 0) {
            historyListEl.innerHTML =
                '<div class="text-center py-8 text-monms-dim text-xs">No messages yet</div>';
            return;
        }

        var html = '';
        for (var i = history.length - 1; i >= 0; i--) {
            var e = history[i];
            var cfg = TYPE_CFG[e.type] || TYPE_CFG.info;
            var hasDetail = !!(e.detail || e.trace);
            html +=
                '<div class="rounded-none ' + cfg.bg + ' ' + cfg.border + ' border group overflow-hidden">' +
                '  <div class="flex items-start gap-2 px-3 py-2 cursor-pointer hover:brightness-110 transition-all" ' +
                     'onclick="window._monmsMsgToggle(this)" data-index="' + i + '" title="' + (hasDetail ? 'Click to expand and copy detail' : 'Click to copy') + '">' +
                '    <span class="material-symbols-outlined text-[14px] mt-px shrink-0 ' + cfg.text + '">' + cfg.icon + '</span>' +
                '    <span class="font-mono text-[10px] opacity-50 mt-px shrink-0">' + esc(e.time) + '</span>' +
                '    <span class="text-xs ' + cfg.text + ' flex-1 break-all select-all">' + esc(e.text) + '</span>' +
                '    <span class="material-symbols-outlined text-[12px] mt-px shrink-0">' + (hasDetail ? 'expand_more' : 'content_copy') + '</span>' +
                '  </div>' +
                (hasDetail
                    ? '<div class="hidden border-t border-monms-border/40 bg-black/20 px-3 py-3 space-y-3" data-role="detail">' +
                      (e.detail ? '<div><div class="text-[10px] uppercase tracking-[0.18em] text-monms-dim mb-2">Detail</div><pre class="text-xs text-monms-text whitespace-pre-wrap break-all">' + esc(e.detail) + '</pre></div>' : '') +
                      (e.trace ? '<div><div class="text-[10px] uppercase tracking-[0.18em] text-monms-dim mb-2">Trace</div><pre class="text-xs text-monms-muted whitespace-pre-wrap break-all">' + esc(formatTrace(e.trace)) + '</pre></div>' : '') +
                      '<div class="text-[10px] text-monms-dim uppercase tracking-[0.18em]">Copied to clipboard on click</div>' +
                    '</div>'
                    : '') +
                '</div>';
        }
        historyListEl.innerHTML = html;
    }

    window._monmsMsgToggle = function (el) {
        var index = parseInt(el.getAttribute('data-index') || '-1', 10);
        if (index < 0 || !history[index]) return;
        var entry = history[index];
        var detailEl = el.parentElement.querySelector('[data-role="detail"]');
        if (detailEl) {
            detailEl.classList.toggle('hidden');
            var icon = el.querySelector('.material-symbols-outlined:last-child');
            if (icon) {
                icon.textContent = detailEl.classList.contains('hidden') ? 'expand_more' : 'expand_less';
            }
        }
        copyText(buildCopyText(entry), function () {
            var orig = el.style.outline;
            el.style.outline = '1px solid #0df2f2';
            setTimeout(function () { el.style.outline = orig; }, 400);
        });
    };

    function openModal() {
        ensureModal();
        renderHistory();
        modalEl.style.display = 'flex';
        modalEl.focus();
    }

    function closeModal() {
        if (modalEl) modalEl.style.display = 'none';
    }

    function renderEmptyStrip() {
        ensureStrip();
        stripEl.className = 'flex items-center gap-2 px-3 py-1.5 text-[10px] cursor-pointer select-none border border-monms-border/40 bg-monms-surface-2/50 text-monms-dim min-w-0 overflow-hidden';
        stripEl.innerHTML =
            '<span class="material-symbols-outlined text-[14px] shrink-0">notifications_none</span>' +
            '<span class="opacity-60 truncate">No messages</span>';
        stripEl.onclick = openModal;
        stripEl.title = 'Click to view message history';
    }

    var confirmStripEl = null;
    var confirmCallback = null;
    var cancelCallback = null;

    function ensureConfirmStrip() {
        if (confirmStripEl) return;
        confirmStripEl = document.createElement('div');
        confirmStripEl.id = 'monms-confirm-strip';
        confirmStripEl.style.display = 'none';
        document.body.appendChild(confirmStripEl);
    }

    function renderConfirmStrip(message) {
        ensureStrip();
        ensureConfirmStrip();
        var cfg = TYPE_CFG.warn;

        confirmStripEl.className = 'fixed top-0 left-0 right-0 flex items-center gap-3 px-4 py-3 text-xs border-b ' +
            cfg.bg + ' ' + cfg.border + ' ' + cfg.text + ' z-50 shadow-lg';
        confirmStripEl.innerHTML =
            '<span class="material-symbols-outlined text-[16px] shrink-0">' + cfg.icon + '</span>' +
            '<span class="flex-1">' + esc(message) + '</span>' +
            '<button id="monms-confirm-yes" class="px-3 py-1 bg-emerald-900/60 hover:bg-emerald-800/60 text-emerald-300 text-[10px] font-bold transition-colors">Confirm</button>' +
            '<button id="monms-confirm-no" class="px-3 py-1 bg-red-900/60 hover:bg-red-800/60 text-red-300 text-[10px] font-bold transition-colors">Cancel</button>';

        var appContainer = document.querySelector('.relative.z-10');
        if (appContainer) {
            appContainer.style.filter = 'blur(5px)';
            appContainer.style.pointerEvents = 'none';
        }

        confirmStripEl.style.animation = 'none';
        setTimeout(function () {
            confirmStripEl.style.animation = 'monms-flash-white 0.4s ease-out';
        }, 10);

        confirmStripEl.style.display = 'block';

        document.getElementById('monms-confirm-yes').onclick = function () {
            if (confirmCallback) confirmCallback();
            closeConfirm();
        };

        document.getElementById('monms-confirm-no').onclick = function () {
            if (cancelCallback) cancelCallback();
            closeConfirm();
        };

        document.addEventListener('keydown', handleConfirmEsc);
    }

    function handleConfirmEsc(e) {
        if (e.key === 'Escape' && confirmStripEl && confirmStripEl.style.display !== 'none') {
            closeConfirm();
        }
    }

    function closeConfirm() {
        if (confirmStripEl) confirmStripEl.style.display = 'none';
        var appContainer = document.querySelector('.relative.z-10');
        if (appContainer) {
            appContainer.style.filter = 'none';
            appContainer.style.pointerEvents = 'auto';
        }
        document.removeEventListener('keydown', handleConfirmEsc);
        confirmCallback = null;
        cancelCallback = null;
    }

    function addMessage(text, type, meta) {
        type = type || 'info';
        if (!TYPE_CFG[type]) type = 'info';
        var normalized = normalizeMeta(meta);
        var entry = {
            text: String(text),
            type: type,
            time: ts(),
            detail: normalized.detail,
            trace: normalized.trace,
            copyText: normalized.copyText
        };
        history.push(entry);
        if (history.length > MAX_HISTORY) history.shift();
        renderStrip(entry);
    }

    function consumeServerFlash() {
        ensureStrip();
        var text = stripEl.getAttribute('data-flash');
        if (!text) return;
        var type = stripEl.getAttribute('data-flash-type') || 'info';
        stripEl.removeAttribute('data-flash');
        stripEl.removeAttribute('data-flash-type');
        addMessage(text, type);
    }

    function init() {
        renderEmptyStrip();
        consumeServerFlash();
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    window.monms = window.monms || {};
    window.monms.msg = addMessage;
    window.monms.msgOk = function (t, meta) { addMessage(t, 'success', meta); };
    window.monms.msgErr = function (t, meta) { addMessage(t, 'error', meta); };
    window.monms.msgWarn = function (t, meta) { addMessage(t, 'warn', meta); };
    window.monms.msgInfo = function (t, meta) { addMessage(t, 'info', meta); };
    window.monms.msgClear = function () {
        history = [];
        renderEmptyStrip();
    };
    window.monms.confirm = function (message, onYes, onNo) {
        confirmCallback = onYes;
        cancelCallback = onNo;
        renderConfirmStrip(message);
    };
    window.monms.esc = esc;
})();
