module.exports = function (api) {
  api.cache(true);
  return {
    presets: [['babel-preset-expo', { jsxImportSource: 'react' }]],
    plugins: [
      // Worklets plugin MUST be last (Reanimated 4 moved it here)
      'react-native-worklets/plugin',
    ],
  };
};
