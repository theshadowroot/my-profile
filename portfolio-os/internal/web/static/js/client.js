document.addEventListener("DOMContentLoaded", () => {
    const nav = document.querySelector(".client-nav");
    const toggle = document.querySelector(".client-menu-toggle");
    const mobileMenu = document.querySelector(".client-mobile-menu");

    if (!nav || !toggle || !mobileMenu) {
        return;
    }

    const mobileLinks = mobileMenu.querySelectorAll("a");

    const setMenuState = (open) => {
        nav.classList.toggle("menu-open", open);

        toggle.setAttribute(
            "aria-expanded",
            String(open)
        );

        toggle.setAttribute(
            "aria-label",
            open
                ? "Close navigation menu"
                : "Open navigation menu"
        );

        mobileMenu.setAttribute(
            "aria-hidden",
            String(!open)
        );
    };

    toggle.addEventListener("click", () => {
        const isOpen = nav.classList.contains("menu-open");

        setMenuState(!isOpen);
    });

    mobileLinks.forEach((link) => {
        link.addEventListener("click", () => {
            setMenuState(false);
        });
    });

    document.addEventListener("click", (event) => {
        if (!nav.contains(event.target)) {
            setMenuState(false);
        }
    });

    document.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            setMenuState(false);
        }
    });

    window.addEventListener("resize", () => {
        if (window.innerWidth > 760) {
            setMenuState(false);
        }
    });
});


// Certificate Lightbox Handler
function openCertModal(imageSrc, title, issuer) {
  const modal = document.getElementById('cert-lightbox-modal');
  const modalImg = document.getElementById('cert-modal-img');
  const modalTitle = document.getElementById('cert-modal-title');
  const modalIssuer = document.getElementById('cert-modal-issuer');

  if (!modal || !modalImg) return;

  modalImg.src = imageSrc;
  modalTitle.textContent = title || '';
  modalIssuer.textContent = issuer || '';

  modal.classList.add('active');
  document.body.style.overflow = 'hidden';
}

function closeCertModalDirect() {
  const modal = document.getElementById('cert-lightbox-modal');
  if (modal) {
    modal.classList.remove('active');
    document.body.style.overflow = '';
  }
}

function closeCertModal(e) {
  if (e.target.id === 'cert-lightbox-modal') {
    closeCertModalDirect();
  }
}

// ESC Key listener to close modal
document.addEventListener('keydown', function (e) {
  if (e.key === 'Escape') {
    closeCertModalDirect();
  }
});