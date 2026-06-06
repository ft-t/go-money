export const DEMO_HOSTNAME = 'gomoney.ft-t.dev';
export const DEMO_CREDENTIALS = { login: 'demo', password: 'demo4vcxsdfss231' };
export const DEMO_PROJECT_URL = 'https://github.com/ft-t/go-money';

export function isDemo(): boolean {
    return localStorage.getItem('is_demo') === 'true' || window.location.hostname === DEMO_HOSTNAME;
}
