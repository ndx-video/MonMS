(function () {
  const AUTH_KEYS = ['pocketbase_auth', '__pb_superusers__'];

  function readToken() {
    for (const key of AUTH_KEYS) {
      try {
        const raw = localStorage.getItem(key);
        if (!raw) continue;
        const parsed = JSON.parse(raw);
        if (parsed && parsed.token) return parsed.token;
      } catch (_) {}
    }
    return '';
  }

  window.monmsSyncAuth = async function () {
    const token = readToken();
    if (!token) return false;
    const res = await fetch('/_monms/auth/sync', {
      method: 'POST',
      headers: { Authorization: 'Bearer ' + token },
    });
    return res.ok || res.status === 204;
  };
})();
