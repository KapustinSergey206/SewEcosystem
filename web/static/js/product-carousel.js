document.addEventListener("DOMContentLoaded", () => {
  document.querySelectorAll(".product-carousel").forEach(initProductCarousel);
});

function initProductCarousel(carousel) {
  const track = carousel.querySelector(".product-carousel-track");
  const slides = carousel.querySelectorAll(".product-carousel-slide");
  const prevBtn = carousel.querySelector(".product-carousel-prev");
  const nextBtn = carousel.querySelector(".product-carousel-next");
  const dotsWrap = carousel.querySelector(".product-carousel-dots");

  if (!track || slides.length === 0) {
    return;
  }

  if (slides.length === 1) {
    carousel.classList.add("product-carousel-single");
    return;
  }

  let current = 0;

  slides.forEach((_, index) => {
    const dot = document.createElement("button");
    dot.type = "button";
    dot.className = "product-carousel-dot" + (index === 0 ? " active" : "");
    dot.setAttribute("aria-label", "Фото " + (index + 1));
    dot.addEventListener("click", (event) => {
      event.stopPropagation();
      showSlide(index);
    });
    dotsWrap.appendChild(dot);
  });

  function showSlide(index) {
    current = index;
    track.style.transform = "translateX(-" + current * 100 + "%)";
    dotsWrap.querySelectorAll(".product-carousel-dot").forEach((dot, i) => {
      dot.classList.toggle("active", i === current);
    });
  }

  prevBtn.addEventListener("click", (event) => {
    event.stopPropagation();
    showSlide((current - 1 + slides.length) % slides.length);
  });

  nextBtn.addEventListener("click", (event) => {
    event.stopPropagation();
    showSlide((current + 1) % slides.length);
  });

  showSlide(0);
}

function buildModalCarousel(images) {
  const track = document.getElementById("modalCarouselTrack");
  const dotsWrap = document.getElementById("modalCarouselDots");
  const prevBtn = document.getElementById("modalCarouselPrev");
  const nextBtn = document.getElementById("modalCarouselNext");
  const carousel = document.getElementById("modalCarousel");

  track.innerHTML = "";
  dotsWrap.innerHTML = "";

  images.forEach((src) => {
    const slide = document.createElement("div");
    slide.className = "product-carousel-slide";
    const img = document.createElement("img");
    img.src = src;
    img.alt = "";
    slide.appendChild(img);
    track.appendChild(slide);
  });

  if (images.length <= 1) {
    carousel.classList.add("product-carousel-single");
    return;
  }

  carousel.classList.remove("product-carousel-single");

  let current = 0;
  const slides = track.querySelectorAll(".product-carousel-slide");

  images.forEach((_, index) => {
    const dot = document.createElement("button");
    dot.type = "button";
    dot.className = "product-carousel-dot" + (index === 0 ? " active" : "");
    dot.setAttribute("aria-label", "Фото " + (index + 1));
    dot.addEventListener("click", () => showSlide(index));
    dotsWrap.appendChild(dot);
  });

  function showSlide(index) {
    current = index;
    track.style.transform = "translateX(-" + current * 100 + "%)";
    dotsWrap.querySelectorAll(".product-carousel-dot").forEach((dot, i) => {
      dot.classList.toggle("active", i === current);
    });
  }

  prevBtn.onclick = () => showSlide((current - 1 + slides.length) % slides.length);
  nextBtn.onclick = () => showSlide((current + 1) % slides.length);
  showSlide(0);
}

function parseProductImages(raw) {
  if (!raw) {
    return [];
  }
  return raw.split("|").map((item) => item.trim()).filter(Boolean);
}
