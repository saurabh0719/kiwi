document.addEventListener('DOMContentLoaded', () => {
    // Highlight active navigation item based on scroll position
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('nav ul li a');
    
    function highlightNavLink() {
        let scrollPosition = window.scrollY;
        
        sections.forEach(section => {
            const sectionTop = section.offsetTop - 100;
            const sectionHeight = section.offsetHeight;
            const sectionId = section.getAttribute('id');
            
            if (scrollPosition >= sectionTop && scrollPosition < sectionTop + sectionHeight) {
                navLinks.forEach(link => {
                    link.classList.remove('active');
                    if (link.getAttribute('href') === `#${sectionId}`) {
                        link.classList.add('active');
                    }
                });
            }
        });
    }
    
    window.addEventListener('scroll', highlightNavLink);
    
    // Add a mobile menu toggle functionality
    const header = document.querySelector('header');
    const nav = document.querySelector('nav');
    
    // Create mobile menu toggle button
    const mobileMenuToggle = document.createElement('button');
    mobileMenuToggle.classList.add('mobile-menu-toggle');
    mobileMenuToggle.innerHTML = '<i class="fas fa-bars"></i>';
    mobileMenuToggle.setAttribute('aria-label', 'Toggle navigation menu');
    
    // Mobile menu functionality
    let menuOpen = false;
    
    function toggleMobileMenu() {
        menuOpen = !menuOpen;
        nav.classList.toggle('active', menuOpen);
        mobileMenuToggle.innerHTML = menuOpen ? 
            '<i class="fas fa-times"></i>' : 
            '<i class="fas fa-bars"></i>';
        document.body.classList.toggle('menu-open', menuOpen);
    }
    
    mobileMenuToggle.addEventListener('click', toggleMobileMenu);
    
    // Only add the toggle for mobile devices
    function updateMobileMenu() {
        if (window.innerWidth <= 768) {
            if (!header.querySelector('.mobile-menu-toggle')) {
                header.querySelector('.header-content').appendChild(mobileMenuToggle);
                nav.classList.add('mobile-nav');
            }
        } else {
            if (header.querySelector('.mobile-menu-toggle')) {
                header.querySelector('.mobile-menu-toggle').remove();
                nav.classList.remove('mobile-nav', 'active');
                document.body.classList.remove('menu-open');
                menuOpen = false;
            }
        }
    }
    
    // Set up initial state and respond to window resizing
    updateMobileMenu();
    window.addEventListener('resize', updateMobileMenu);
    
    // Close mobile menu when clicking on links
    navLinks.forEach(link => {
        link.addEventListener('click', () => {
            if (menuOpen) {
                toggleMobileMenu();
            }
        });
    });
    
    // Smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            e.preventDefault();
            
            const targetId = this.getAttribute('href');
            if (targetId === '#') return;
            
            const targetElement = document.querySelector(targetId);
            if (targetElement) {
                window.scrollTo({
                    top: targetElement.offsetTop - 70,
                    behavior: 'smooth'
                });
            }
        });
    });
    
    // Add active class to navigation links
    highlightNavLink();
    
    // Animate feature cards on scroll
    const featureCards = document.querySelectorAll('.feature-card');
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('animated');
                observer.unobserve(entry.target);
            }
        });
    }, {
        threshold: 0.2
    });
    
    featureCards.forEach(card => {
        observer.observe(card);
    });
}); 