export interface QuickTag {
    label: string;
    tagIds: number[];
}

export interface AccountsListConfig {
    quickTags: QuickTag[];
}

export const ACCOUNTS_LIST_PAGE_ID = 'accounts-list';

export const ACCOUNTS_LIST_DEFAULTS: AccountsListConfig = Object.freeze({
    quickTags: Object.freeze([] as QuickTag[]),
}) as unknown as AccountsListConfig;
