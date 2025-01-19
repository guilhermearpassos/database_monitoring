
function openSlideover() {
    document.getElementById('slideover-wrapper').classList.add('open');
}

function closeSlideover() {
    document.getElementById('slideover-wrapper').classList.remove('open');
}

function handleOutsideClick(event) {
    const slideover = document.getElementById('slideover');
    if (!slideover.contains(event.target)) {
        closeSlideover();
    }
}