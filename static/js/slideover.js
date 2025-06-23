function openSlideover() {
    document.getElementById('slideover-wrapper').classList.add('open');
}

function closeSlideover() {
    // Navigate back to home page with proper URL handling
    htmx.ajax('GET', '/', {
        target: 'body',
        swap: 'outerHTML'
    }).then(() => {
        window.history.pushState({}, '', '/');
    });
}

function handleOutsideClick(event) {
    const slideover = document.getElementById('slideover');
    if (!slideover.contains(event.target)) {
        closeSlideover();
    }
}

// Handle browser back button
window.addEventListener('popstate', function(event) {
    const currentPath = window.location.pathname;
    
    if (currentPath === '/') {
        // Close slideover if we're back to home
        const slideoverWrapper = document.getElementById('slideover-wrapper');
        if (slideoverWrapper && slideoverWrapper.classList.contains('open')) {
            slideoverWrapper.classList.remove('open');
        }
    }
});