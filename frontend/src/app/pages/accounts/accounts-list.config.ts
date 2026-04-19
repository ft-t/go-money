export interface QuickTag {
    label: string;
    search: string;
}

export interface AccountsListConfig {
    quickTags: QuickTag[];
}

export const ACCOUNTS_LIST_PAGE_ID = 'accounts-list';

export const ACCOUNTS_LIST_DEFAULTS: AccountsListConfig = {
    quickTags: [],
};
