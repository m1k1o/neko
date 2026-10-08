// the two React hook rules; types and formatting are checked by tsc and prettier
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'

export default [
  { ignores: ['dist', 'core/src/keyboard/guacamole.js'] },
  {
    files: ['src/**/*.{ts,tsx}', 'core/src/**/*.ts'],
    languageOptions: { parser: tseslint.parser },
    plugins: { 'react-hooks': reactHooks },
    rules: { 'react-hooks/rules-of-hooks': 'error', 'react-hooks/exhaustive-deps': 'error' },
  },
]
