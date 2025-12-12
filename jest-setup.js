// Jest setup provided by Grafana scaffolding
import 'jest-canvas-mock';
import './.config/jest-setup';

// Mock canvas context methods for Combobox component
Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
  value: () => ({
    measureText: () => ({ width: 100 }),
    fillText: () => {},
    clearRect: () => {},
    getImageData: () => ({
      data: new Array(4).fill(255)
    }),
    putImageData: () => {},
    createImageData: () => ({
      data: new Array(4).fill(255)
    }),
    setTransform: () => {},
    drawImage: () => {},
    save: () => {},
    fillRect: () => {},
    restore: () => {},
    beginPath: () => {},
    moveTo: () => {},
    lineTo: () => {},
    closePath: () => {},
    stroke: () => {},
    translate: () => {},
    scale: () => {},
    rotate: () => {},
    arc: () => {},
    fill: () => {},
    transform: () => {},
    rect: () => {},
    clip: () => {},
  }),
});

// Mock getBoundingClientRect for better component testing
Element.prototype.getBoundingClientRect = jest.fn(() => ({
  width: 100,
  height: 20,
  top: 0,
  left: 0,
  bottom: 20,
  right: 100,
  x: 0,
  y: 0,
  toJSON: () => {},
}));
