// Patches RN 0.81's jest/mockComponent.js, which crashes under React 19 because
// some RN components are React.forwardRef objects whose `.prototype` is undefined.
// We replace the SuperClass-detection branch with a null-safe variant.

const Module = require('module');
const path = require('path');
const fs = require('fs');

const originalLoad = Module._load;
Module._load = function patchedLoad(request, parent, isMain) {
  if (typeof request === 'string' && request.endsWith('/jest/mockComponent')) {
    const resolved = Module._resolveFilename(request, parent, isMain);
    if (resolved.endsWith('mockComponent.js')) {
      return require(resolved);
    }
  }
  return originalLoad.apply(this, arguments);
};

const rnRoot = (() => {
  try {
    return path.dirname(require.resolve('react-native/package.json'));
  } catch {
    return null;
  }
})();

if (rnRoot) {
  const mockComponentPath = path.join(rnRoot, 'jest', 'mockComponent.js');
  try {
    const src = fs.readFileSync(mockComponentPath, 'utf8');
    if (src.includes('RealComponent.prototype.constructor instanceof React.Component')) {
      const patched = src.replace(
        'typeof RealComponent === \'function\' &&\n    RealComponent.prototype.constructor instanceof React.Component',
        'typeof RealComponent === \'function\' &&\n    RealComponent.prototype != null &&\n    RealComponent.prototype.constructor instanceof React.Component',
      );
      if (patched !== src) {
        fs.writeFileSync(mockComponentPath, patched);
      }
    }
  } catch {
    // ignore — patch is best-effort
  }
}
