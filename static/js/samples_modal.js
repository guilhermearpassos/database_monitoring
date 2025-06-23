function openSlideover() {
    document.getElementById('samples_modal_container').classList.add('open');
    
    // Add ESC key listener
    document.addEventListener('keydown', handleEscapeKey);
    
    // Prevent body scrolling
    document.body.style.overflow = 'hidden';
}

function closeSampleModal() {
    document.getElementById('samples_modal_container').classList.remove('open');
    
    // Remove ESC key listener
    document.removeEventListener('keydown', handleEscapeKey);
    
    // Restore body scrolling
    document.body.style.overflow = '';
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