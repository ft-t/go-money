// rabbit
export class CustomEnumError {
  static None = 0;
  static DuplicateKeyValue = 2;
  static MissingJwtToken = 13;
  static ExpiredJwtToken = 14;
  static InvalidJwtToken = 15;
  static RecordNotFound = 19;

  static transformKnownErrorCodes(errorCode: number): string | null {
    switch (errorCode) {
      default:
        return null;
    }
  }
}
