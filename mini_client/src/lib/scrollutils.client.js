// https://stackoverflow.com/a/4770179

function preventDefault(e) {
  e.preventDefault();
}

// modern Chrome requires { passive: false } when adding event
var supportsPassive = false;
try {
  window.addEventListener("test", null, Object.defineProperty({}, 'passive', {
    get: function () { supportsPassive = true; } 
  }));
} catch(e) {}

var wheelOpt = supportsPassive ? { passive: false } : false;
var wheelEvent = 'onwheel' in document.createElement('div') ? 'wheel' : 'mousewheel';

// call this to Disable
export function disableTouchScroll() {
  window.addEventListener('touchmove', preventDefault, wheelOpt); // mobile
}

// call this to Enable
export function enableTouchScroll() {
  window.removeEventListener('touchmove', preventDefault);
}