export interface SnippetTransaction {
    type: number;
    title: string;
    notes: string;
    sourceAccountId: number;
    sourceCurrency: string;
    sourceAmount: string;
    destinationAccountId: number;
    destinationCurrency: string;
    destinationAmount: string;
    categoryId: number;
    tagIds: number[];
    fxSourceAmount: string;
    fxSourceCurrency: string;
    internalReferenceNumbers: string[];
}

export interface Snippet {
    id: string;
    name: string;
    transactions: SnippetTransaction[];
    createdAt: string;
    updatedAt: string;
}
