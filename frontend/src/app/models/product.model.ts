export interface Product {
  CellLocation: string;
  ReceivedDate: Date;
  IssuedDate: Date | null;
  ProductName: string;
  ReceivedByEmployee: string;
  IssuedByEmployee: string | null;
  WeightReceived: number;
  WeightIssued: number | null;
  Barcode?: string;
  StorageCell?: string;
  ProductBarcode?: string;
  ArticleNumber?: string;
}

export enum SearchType {
  BARCODE = 'barcode',
  STORAGE_CELL = 'storageCell',
  PRODUCT_BARCODE = 'productBarcode',
  ARTICLE = 'article'
}

export interface SearchParams {
  searchType: SearchType;
  searchValue: string;
}