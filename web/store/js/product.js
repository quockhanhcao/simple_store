// product.js: Views/Catalog/Details.cshtml, at /product.html?id=N.

import { api, addToCart } from "./api.js";
import { esc, formatPrice, flashSuccess, showError } from "./layout.js";

const el = document.getElementById("product");
const id = new URLSearchParams(location.search).get("id") ?? "";

function render(p) {
  document.title = `${p.name} - SimpleStore`;
  el.innerHTML = `
    <nav aria-label="breadcrumb" class="mb-4">
      <ol class="breadcrumb">
        <li class="breadcrumb-item"><a href="/">Catalog</a></li>
        <li class="breadcrumb-item"><a href="/?categoryId=${p.categoryId}">${esc(p.category.name)}</a></li>
        <li class="breadcrumb-item active">${esc(p.name)}</li>
      </ol>
    </nav>
    <div class="row g-4">
      <div class="col-md-5">
        <div class="card shadow-sm">
          <div class="card-body d-flex align-items-center justify-content-center bg-light" style="height:400px;">
            <i class="bi bi-box-seam" style="font-size:8rem; color:#adb5bd;"></i>
          </div>
        </div>
      </div>
      <div class="col-md-7">
        <span class="badge bg-secondary mb-2">${esc(p.category.name)}</span>
        <h1 class="mb-3">${esc(p.name)}</h1>
        <p class="lead text-muted">${esc(p.description)}</p>
        <div class="fs-2 fw-bold text-primary mb-3">${formatPrice(p.price)}</div>
        <p class="text-muted"><i class="bi bi-check-circle-fill text-success me-2"></i>${p.stock} items in stock</p>
        <hr />
        <form id="add-to-cart" class="d-flex gap-3 align-items-center">
          <div class="input-group" style="width:120px;">
            <span class="input-group-text">Qty</span>
            <input type="number" name="quantity" value="1" min="1" max="${p.stock}" class="form-control" />
          </div>
          <button type="submit" class="btn btn-primary btn-lg flex-grow-1">
            <i class="bi bi-cart-plus me-2"></i>Add to Cart
          </button>
        </form>
        <a href="/" class="btn btn-outline-secondary mt-3">
          <i class="bi bi-arrow-left me-2"></i>Back to Catalog
        </a>
      </div>
    </div>`;

  document.getElementById("add-to-cart").addEventListener("submit", async (e) => {
    e.preventDefault();
    const form = e.currentTarget;
    const button = form.querySelector("button");
    button.disabled = true;
    try {
      await addToCart(p.id, Number(form.quantity.value));
      flashSuccess("Item added to cart!");
      location.href = "/";
    } catch (err) {
      showError(err.message);
      button.disabled = false;
    }
  });
}

function renderNotFound() {
  document.title = "Not Found - SimpleStore";
  el.innerHTML = `
    <div class="text-center py-5">
      <i class="bi bi-search display-1 text-muted"></i>
      <h3 class="mt-3">Product not found</h3>
      <a href="/" class="btn btn-primary mt-2">Back to Catalog</a>
    </div>`;
}

async function load() {
  try {
    render(await api("/products/" + encodeURIComponent(id)));
  } catch (err) {
    // C# returns 404 for a missing product and for a bad id (which binds to 0).
    if (err.status === 404 || err.status === 400) {
      renderNotFound();
    } else {
      showError(err.message);
    }
  }
}

load();
