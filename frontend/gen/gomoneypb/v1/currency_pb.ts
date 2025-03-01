// @generated by protoc-gen-es v1.10.0 with parameter "target=ts,import_extension=none"
// @generated from file gomoneypb/v1/currency.proto (package gomoneypb.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from message gomoneypb.v1.Currency
 */
export class Currency extends Message<Currency> {
  /**
   * @generated from field: string id = 1;
   */
  id = "";

  /**
   * @generated from field: string rate = 2;
   */
  rate = "";

  /**
   * @generated from field: bool is_active = 3;
   */
  isActive = false;

  /**
   * @generated from field: int32 decimal_places = 4;
   */
  decimalPlaces = 0;

  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 5;
   */
  updatedAt?: Timestamp;

  /**
   * @generated from field: google.protobuf.Timestamp deleted_at = 6;
   */
  deletedAt?: Timestamp;

  constructor(data?: PartialMessage<Currency>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "gomoneypb.v1.Currency";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "rate", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "is_active", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 4, name: "decimal_places", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 5, name: "updated_at", kind: "message", T: Timestamp },
    { no: 6, name: "deleted_at", kind: "message", T: Timestamp },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Currency {
    return new Currency().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Currency {
    return new Currency().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Currency {
    return new Currency().fromJsonString(jsonString, options);
  }

  static equals(a: Currency | PlainMessage<Currency> | undefined, b: Currency | PlainMessage<Currency> | undefined): boolean {
    return proto3.util.equals(Currency, a, b);
  }
}

