// catalog.js: Views/Catalog/Index.cshtml.
// The page state lives in the URL (?categoryId=&search=), as in the C# app,
// and both values are passed to the API unchanged. The API decides what they mean.

import { api, addToCart } from "./api.js";
import { esc, formatPrice, flashSuccess, showError } from "./layout.js";

const params = new URLSearchParams(location.search);
const categoryId = params.get("categoryId");
const search = params.get("search") ?? "";

document.getElementById("search").value = search;

function renderCategories(categories) {
  const list = document.getElementById("categories");
  document.getElementById("all-products").classList.toggle("active", !categoryId);
  for (const cat of categories) {
    list.insertAdjacentHTML(
      "beforeend",
      `<a href="/?categoryId=${cat.id}" class="list-group-item list-group-item-action ${String(cat.id) === categoryId ? "active" : ""}">
        <i class="bi bi-tag me-2"></i>${esc(cat.name)}
      </a>`,
    );
  }
}

function productCard(p) {
  return `
    <div class="col">
      <div class="card h-100 shadow-sm product-card">
        <div class="product-image-container bg-light d-flex align-items-center justify-content-center" style="height:200px; overflow:hidden;">
          <i class="bi bi-box-seam display-1 text-secondary"></i>
        </div>
        <div class="card-body d-flex flex-column">
          <span class="badge bg-secondary mb-2 align-self-start">${esc(p.category.name)}</span>
          <h5 class="card-title">${esc(p.name)}</h5>
          <p class="card-text text-muted small flex-grow-1">${esc(p.description)}</p>
          <div class="d-flex justify-content-between align-items-center mt-2">
            <span class="fs-4 fw-bold text-primary">${formatPrice(p.price)}</span>
            <span class="text-muted small">${p.stock} in stock</span>
          </div>
        </div>
        <div class="card-footer bg-transparent border-top-0 pb-3">
          <div class="d-grid gap-2">
            <a href="/product.html?id=${p.id}" class="btn btn-outline-primary">
              <i class="bi bi-eye me-1"></i>View Details
            </a>
            <form data-add-to-cart data-product-id="${p.id}">
              <button type="submit" class="btn btn-primary w-100">
                <i class="bi bi-cart-plus me-1"></i>Add to Cart
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>`;
}

function renderProducts(products) {
  const el = document.getElementById("products");
  if (products.length === 0) {
    el.innerHTML = `
      <div class="text-center py-5">
        <i class="bi bi-search display-1 text-muted"></i>
        <h3 class="mt-3">No products found</h3>
        <a href="/" class="btn btn-primary mt-2">View All Products</a>
      </div>`;
    return;
  }
  el.innerHTML = `<div class="row row-cols-1 row-cols-md-2 row-cols-xl-3 g-4">${products.map(productCard).join("")}</div>`;
}

// C# posts the form, then redirects to the catalog with "Item added to cart!".
document.getElementById("products").addEventListener("submit", async (e) => {
  const form = e.target.closest("[data-add-to-cart]");
  if (!form) return;
  e.preventDefault();
  const button = form.querySelector("button");
  button.disabled = true;
  try {
    await addToCart(Number(form.dataset.productId), 1);
    flashSuccess("Item added to cart!");
    location.href = "/";
  } catch (err) {
    showError(err.message);
    button.disabled = false;
  }
});

async function load() {
  const query = new URLSearchParams();
  if (categoryId !== null) query.set("categoryId", categoryId);
  if (search !== "") query.set("search", search);
  const qs = query.toString();

  try {
    const [categories, products] = await Promise.all([
      api("/categories"),
      api("/products" + (qs ? "?" + qs : "")),
    ]);
    renderCategories(categories);
    renderProducts(products);
  } catch (err) {
    showError(err.message);
  }
}

load();
