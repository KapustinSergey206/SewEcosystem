const products = document.querySelectorAll(".calc-product");

const searchInput = document.getElementById("productSearch");

const prevBtn = document.getElementById("prevPage");
const nextBtn = document.getElementById("nextPage");

const pageInfo = document.getElementById("pageInfo");

const qtyInputs = document.querySelectorAll(".qty-input");

const totalPrice = document.getElementById("totalPrice");
const itemsCount = document.getElementById("itemsCount");

const perPage = 6;

let currentPage = 1;

function calculateTotal() {

    let total = 0;
    let items = 0;

    qtyInputs.forEach(input => {

        const qty = parseInt(input.value) || 0;

        const price = parseFloat(
            input.dataset.price
        );

        if (qty > 0) {

            total += qty * price;

            items += qty;

        }

    });

    itemsCount.innerText = items;

    totalPrice.innerText =
        total.toFixed(2) + " ₽";

}

function renderProducts() {

    let filtered = [...products].filter(product => {

        const title = product
            .querySelector("h3")
            .innerText
            .toLowerCase();

        return title.includes(
            searchInput.value.toLowerCase()
        );

    });

    const totalPages =
        Math.ceil(filtered.length / perPage);

    if (currentPage > totalPages) {
        currentPage = 1;
    }

    products.forEach(p => {
        p.style.display = "none";
    });

    filtered
        .slice(
            (currentPage - 1) * perPage,
            currentPage * perPage
        )
        .forEach(p => {
            p.style.display = "flex";
        });

    pageInfo.innerText =
        currentPage + " / " + totalPages;

}

searchInput.addEventListener("input", () => {

    currentPage = 1;

    renderProducts();

});

nextBtn.addEventListener("click", () => {

    currentPage++;

    renderProducts();

});

prevBtn.addEventListener("click", () => {

    if (currentPage > 1) {

        currentPage--;

    }

    renderProducts();

});

qtyInputs.forEach(input => {

    input.addEventListener("input", () => {

        calculateTotal();

    });

});

calculateTotal();

renderProducts();