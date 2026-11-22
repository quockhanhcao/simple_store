// api.js: same-origin calls to the Go JSON API.
//
// Every non-2xx response carries the error envelope
// {"error":{"code","message","fields"?}}, which becomes an ApiError.

export class ApiError extends Error {
  constructor(status, code, message, fields) {
    super(message);
    this.status = status;
    this.code = code;
    this.fields = fields || {};
  }
}

export async function api(path, { method = "GET", body } = {}) {
  const opts = {
    method,
    headers: { Accept: "application/json" },
    credentials: "same-origin", // the session cookie (Day 6+) rides along
  };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }

  let res;
  try {
    res = await fetch("/api/v1" + path, opts);
  } catch {
    throw new ApiError(0, "network", "Could not reach the server. Please try again.");
  }

  // 401: the user must sign in (Day 7). Send them to login and bring them back here.
  if (res.status === 401) {
    const returnUrl = encodeURIComponent(location.pathname + location.search);
    location.href = "/login.html?returnUrl=" + returnUrl;
    throw new ApiError(401, "unauthorized", "Please sign in.");
  }

  if (res.status === 204) return null;

  let data = null;
  try {
    data = await res.json();
  } catch {
    // not JSON; handled below
  }

  if (!res.ok) {
    const e = data && data.error;
    throw new ApiError(
      res.status,
      (e && e.code) || "unknown",
      (e && e.message) || `Request failed (${res.status}).`,
      e && e.fields,
    );
  }
  return data;
}

// POST /api/v1/cart/items. Until Day 6 this returns the /api/ JSON 404, which the page shows as an error.
export async function addToCart(productId, quantity) {
  await api("/cart/items", { method: "POST", body: { productId, quantity } });
}
