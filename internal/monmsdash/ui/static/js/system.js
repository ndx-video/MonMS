/**
 * System page — confirm restart/shutdown before POST.
 */
(function () {
  'use strict';

  function bindConfirm(formId, message, onYes) {
    var form = document.getElementById(formId);
    if (!form) return;
    form.addEventListener('submit', function (ev) {
      ev.preventDefault();
      if (typeof window.monms !== 'undefined' && typeof window.monms.confirm === 'function') {
        window.monms.confirm(message, function () {
          form.submit();
        });
        return;
      }
      if (window.confirm(message)) {
        form.submit();
      }
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    bindConfirm(
      'monms-system-restart-form',
      'Restart all MonMS processes using this binary? Active connections will drop.',
      null
    );
    bindConfirm(
      'monms-system-shutdown-form',
      'Shut down all MonMS processes using this binary? You will need to start serve again from the shell.',
      null
    );
  });
})();
