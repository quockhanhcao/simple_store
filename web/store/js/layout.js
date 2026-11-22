// layout.js: the shared navbar, footer and flash alerts (Views/Shared/_Layout.cshtml).
// Each page has <div id="navbar"></div>, <main>, and <div id="footer"></div>.
//
// Placeholders until later days:
//   - the cart badge stays hidden until Day 6 (GET /api/v1/cart)
//   - the user menu always shows Sign In / Register until Day 7 (GET /api/v1/me)

const FLASH_KEY = "ss.flash";

// esc makes text safe to put inside HTML built with template strings.
export function esc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

// formatPrice matches C#'s Price.ToString("N2") in en-US: 1,299.99
const priceFormat = new Intl.NumberFormat("en-US", {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});
export function formatPrice(n) {
  return "$" + priceFormat.format(n);
}

// flashSuccess stores a message to show on the next page, like TempData["Success"].
export function flashSuccess(message) {
  try {
    sessionStorage.setItem(FLASH_KEY, message);
  } catch {
    // storage blocked: the message is just skipped
  }
}

function takeFlash() {
  try {
    const msg = sessionStorage.getItem(FLASH_KEY);
    sessionStorage.removeItem(FLASH_KEY);
    return msg;
  } catch {
    return null;
  }
}

function alertHtml(kind, icon, message) {
  return `
    <div class="container mt-3">
      <div class="alert alert-${kind} alert-dismissible fade show">
        <i class="bi ${icon} me-2"></i>${esc(message)}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
      </div>
    </div>`;
}

// showError puts a dismissible red alert at the top of <main>.
export function showError(message) {
  const main = document.querySelector("main");
  main.querySelector("[data-error-alert]")?.remove();
  const wrap = document.createElement("div");
  wrap.dataset.errorAlert = "";
  wrap.innerHTML = alertHtml("danger", "bi-exclamation-triangle", message);
  main.prepend(wrap);
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function renderNavbar() {
  document.getElementById("navbar").outerHTML = `
    <nav class="navbar navbar-expand-lg sticky-top ss-navbar">
      <div class="container">
        <a class="navbar-brand fw-bold" href="/">
          <i class="bi bi-shop me-2"></i>SimpleStore
        </a>
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav me-auto">
            <li class="nav-item">
              <a class="nav-link" href="/">
                <i class="bi bi-grid me-1"></i>Catalog
              </a>
            </li>
          </ul>
          <ul class="navbar-nav">
            <li class="nav-item">
              <a class="nav-link" href="/cart.html">
                <i class="bi bi-cart3 me-1"></i>Cart
                <span id="cart-count" class="badge bg-danger" hidden></span>
              </a>
            </li>
            <li class="nav-item">
              <a class="nav-link" href="/login.html">
                <i class="bi bi-box-arrow-in-right me-1"></i>Sign In
              </a>
            </li>
            <li class="nav-item">
              <a class="nav-link" href="/register.html">
                <i class="bi bi-person-plus me-1"></i>Register
              </a>
            </li>
          </ul>
        </div>
      </div>
    </nav>`;
}

function renderFooter() {
  document.getElementById("footer").outerHTML = `
    <footer class="ss-footer py-4 mt-5">
      <div class="container text-center">
        <p class="mb-0">&copy; ${new Date().getFullYear()} SimpleStore. Built with Go &amp; PostgreSQL.</p>
      </div>
    </footer>`;
}

renderNavbar();
renderFooter();

const flash = takeFlash();
if (flash) {
  document.querySelector("main").insertAdjacentHTML("afterbegin", alertHtml("success", "bi-check-circle", flash));
}
