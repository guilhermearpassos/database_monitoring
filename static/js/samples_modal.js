
function openSlideover() {
    document.getElementById('samples_modal_container').classList.add('open');
}

function closeSampleModal() {
    document.getElementById('samples_modal_container').classList.remove('open');
}

function handleOutsideClickSamplesModal(event) {
    const slideover = document.getElementById('samples_modal');
    if (!slideover.contains(event.target)) {
        closeSampleModal();
    }
}