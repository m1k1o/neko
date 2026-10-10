// The two React hook rules, and the import rules of the layered layout (see README.md): types and
// formatting are checked by tsc and prettier.
import tseslint from 'typescript-eslint'
import reactHooks from 'eslint-plugin-react-hooks'

// Dependencies point one way: app -> layout -> features -> components -> design, with state, i18n
// and lib below everything but design. The app shell and the layout reach plugins only through the
// registry (`@/plugins`); features see only the contract (`@/plugins/types`) and get plugin
// contributions passed in, and plugins import neither features nor each other.
// Between folders only `@/` imports are used (a feature by its main file, `@/features/video/Stage`);
// a folder's own files are imported relatively.
const only = (allowed) => ({
  'no-restricted-imports': [
    'error',
    {
      patterns: [
        { group: ['../*'], message: 'import other layers with @/ (relative imports stay inside the folder)' },
        { group: ['@m1k1o/neko/*', '@/core/*'], message: 'the core is imported from its entry point @m1k1o/neko' },
        { group: ['@/plugins/*', '!@/plugins/types'], message: 'plugins are reached through the registry @/plugins' },
        ...allowed,
      ],
    },
  ],
})
const not = (...layers) => layers.map((l) => ({ group: [`@/${l}`, `@/${l}/*`], message: `${l} is above this layer` }))

export default [
  { ignores: ['dist', 'core/dist', 'core/src/keyboard/guacamole.js'] },
  {
    files: ['src/**/*.{ts,tsx}', 'core/src/**/*.ts'],
    languageOptions: { parser: tseslint.parser },
    plugins: { 'react-hooks': reactHooks },
    rules: { 'react-hooks/rules-of-hooks': 'error', 'react-hooks/exhaustive-deps': 'error' },
  },
  { files: ['src/components/**'], rules: only(not('app', 'layout', 'features', 'plugins')) },
  {
    files: ['src/state/**', 'src/i18n/**', 'src/lib/**'],
    rules: only(not('app', 'layout', 'features', 'plugins', 'components')),
  },
  {
    files: ['src/features/**'],
    rules: only([
      ...not('app', 'layout'),
      // the registry and everything under it but the contract (a group cannot re-include a child of `@/plugins`)
      {
        regex: '^@/plugins(/(?!types$)|$)',
        message: 'features see only the plugin contract @/plugins/types; the app shell passes contributions in',
      },
    ]),
  },
  { files: ['src/plugins/**'], rules: only(not('app', 'layout', 'features')) },
  { files: ['src/layout/**'], rules: only(not('app')) },
  { files: ['src/app/**', 'src/main.tsx'], rules: only([]) },
]
