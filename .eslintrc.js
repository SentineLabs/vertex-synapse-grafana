module.exports = {
  extends: ['@grafana/eslint-config'],
  root: true,
  rules: {
    '@typescript-eslint/no-explicit-any': 'warn',
    '@typescript-eslint/consistent-type-assertions': 'warn',
    'react-hooks/exhaustive-deps': 'warn',
  },
};
