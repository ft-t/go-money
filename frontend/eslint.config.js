// @ts-check
import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import angular from 'angular-eslint';
import prettierPlugin from 'eslint-plugin-prettier';
import prettierConfig from 'eslint-config-prettier';

export default tseslint.config(
  {
    ignores: ['dist/**', 'coverage/**', '.angular/**', 'node_modules/**', 'public/**']
  },
  {
    files: ['**/*.ts'],
    extends: [
      eslint.configs.recommended,
      ...tseslint.configs.recommended,
      ...angular.configs.tsRecommended,
      prettierConfig
    ],
    plugins: {
      prettier: prettierPlugin
    },
    processor: angular.processInlineTemplates,
    rules: {
      '@angular-eslint/component-selector': ['error', {type: 'element', prefix: 'app', style: 'kebab-case'}],
      '@angular-eslint/directive-selector': ['error', {type: 'attribute', prefix: 'app', style: 'camelCase'}],
      '@angular-eslint/component-class-suffix': ['error', {suffixes: ['']}],
      '@angular-eslint/no-host-metadata-property': 'off',
      '@angular-eslint/no-output-on-prefix': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-inferrable-types': 'off',
      '@typescript-eslint/no-unused-vars': ['warn', {argsIgnorePattern: '^_'}],
      '@typescript-eslint/member-ordering': [
        'error',
        {default: ['public-static-field', 'static-field', 'instance-field', 'public-instance-method']}
      ],
      'arrow-body-style': ['error', 'as-needed'],
      'curly': 'off',
      'no-console': 'off',
      'prefer-const': 'off',
      'padding-line-between-statements': [
        'error',
        {blankLine: 'always', prev: ['const', 'let', 'var'], next: '*'},
        {blankLine: 'any', prev: ['const', 'let', 'var'], next: ['const', 'let', 'var']},
        {blankLine: 'any', prev: ['case', 'default'], next: 'break'},
        {blankLine: 'any', prev: 'case', next: 'case'},
        {blankLine: 'always', prev: '*', next: 'return'},
        {blankLine: 'always', prev: 'block', next: '*'},
        {blankLine: 'always', prev: '*', next: 'block'},
        {blankLine: 'always', prev: 'block-like', next: '*'},
        {blankLine: 'always', prev: '*', next: 'block-like'},
        {blankLine: 'always', prev: 'import', next: ['const', 'let', 'var']}
      ]
    }
  },
  {
    files: ['**/*.html'],
    extends: [
      ...angular.configs.templateRecommended,
      ...angular.configs.templateAccessibility
    ],
    rules: {
      '@angular-eslint/template/eqeqeq': ['error', {allowNullOrUndefined: true}]
    }
  }
);
