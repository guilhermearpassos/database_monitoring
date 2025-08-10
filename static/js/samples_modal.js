// samples_modal.js - Optimize memory usage
function openSlideover() {
    const modalContainer = document.getElementById('samples_modal_container');
    if (!modalContainer) return;
    
    modalContainer.classList.add('open');
    
    // Add ESC key listener
    document.addEventListener('keydown', handleEscapeKey);
    
    // Prevent body scrolling
    document.body.style.overflow = 'hidden';
    
    // Optimize memory usage by cleaning up unused data
    setTimeout(cleanupUnusedData, 500);
}

function cleanupUnusedData() {
    // Store SQL text in data attributes and remove from DOM until needed
    document.querySelectorAll('.full-sql-content').forEach(element => {
        const parent = element.closest('.sql-container');
        if (parent) {
            // Store the content in data attribute and remove the element
            parent.dataset.fullSql = element.textContent;
            element.textContent = '';
            element.classList.add('deferred-content');
        }
    });
    
    // Set up lazy loading of SQL content
    document.querySelectorAll('.sql-container').forEach(container => {
        container.addEventListener('mouseenter', function() {
            const deferredEl = this.querySelector('.deferred-content');
            if (deferredEl && this.dataset.fullSql) {
                deferredEl.textContent = this.dataset.fullSql;
                delete this.dataset.fullSql; // Free memory from the attribute
            }
        }, { once: true });
    });
}

// Add cleanup when closing modal
function closeSampleModal() {
    const modalContainer = document.getElementById('samples_modal_container');
    if (modalContainer) {
        modalContainer.classList.remove('open');
    }
    
    // Re-enable scrolling
    document.body.style.overflow = '';
    
    // Remove ESC key listener
    document.removeEventListener('keydown', handleEscapeKey);
    
    // Clean up memory
    document.querySelectorAll('.deferred-content').forEach(element => {
        element.textContent = '';
    });
}

function handleEscapeKey(event) {
    if (event.key === 'Escape') {
        closeSampleModal();
    }
}

function handleOutsideClickSamplesModal(event) {
    // Get the white content area (the actual modal content)
    const modalContent = event.currentTarget.querySelector('.bg-white');
    
    // If the click target is not contained within the white modal content, close the modal
    if (modalContent && !modalContent.contains(event.target)) {
        closeSampleModal();
    }
}