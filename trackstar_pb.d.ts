// @generated by protoc-gen-es v1.10.0
// @generated from file trackstar.proto (package trackstar, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum trackstar.BusTopic
 */
export declare enum BusTopic {
  /**
   * @generated from enum value: TRACKSTAR = 0;
   */
  TRACKSTAR = 0,
}

/**
 * @generated from enum trackstar.MessageType
 */
export declare enum MessageType {
  /**
   * @generated from enum value: TYPE_UNSPECIFIED = 0;
   */
  TYPE_UNSPECIFIED = 0,

  /**
   * @generated from enum value: TYPE_DECK_DISCOVERED = 1;
   */
  TYPE_DECK_DISCOVERED = 1,

  /**
   * @generated from enum value: TYPE_TRACK_UPDATE = 2;
   */
  TYPE_TRACK_UPDATE = 2,

  /**
   * @generated from enum value: TYPE_DECK_STYLE_UPDATE = 3;
   */
  TYPE_DECK_STYLE_UPDATE = 3,
}

/**
 * @generated from message trackstar.DeckDiscovered
 */
export declare class DeckDiscovered extends Message<DeckDiscovered> {
  /**
   * @generated from field: string deck_id = 1;
   */
  deckId: string;

  constructor(data?: PartialMessage<DeckDiscovered>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.DeckDiscovered";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeckDiscovered;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeckDiscovered;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeckDiscovered;

  static equals(a: DeckDiscovered | PlainMessage<DeckDiscovered> | undefined, b: DeckDiscovered | PlainMessage<DeckDiscovered> | undefined): boolean;
}

/**
 * @generated from message trackstar.TrackUpdate
 */
export declare class TrackUpdate extends Message<TrackUpdate> {
  /**
   * @generated from field: string deck_id = 1;
   */
  deckId: string;

  /**
   * @generated from field: trackstar.Track track = 2;
   */
  track?: Track;

  constructor(data?: PartialMessage<TrackUpdate>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.TrackUpdate";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TrackUpdate;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TrackUpdate;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TrackUpdate;

  static equals(a: TrackUpdate | PlainMessage<TrackUpdate> | undefined, b: TrackUpdate | PlainMessage<TrackUpdate> | undefined): boolean;
}

/**
 * @generated from message trackstar.Track
 */
export declare class Track extends Message<Track> {
  /**
   * @generated from field: string artist = 1;
   */
  artist: string;

  /**
   * @generated from field: string title = 2;
   */
  title: string;

  constructor(data?: PartialMessage<Track>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.Track";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Track;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Track;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Track;

  static equals(a: Track | PlainMessage<Track> | undefined, b: Track | PlainMessage<Track> | undefined): boolean;
}

/**
 * @generated from message trackstar.DeckStyleUpdate
 */
export declare class DeckStyleUpdate extends Message<DeckStyleUpdate> {
  /**
   * @generated from field: string deck_id = 1;
   */
  deckId: string;

  /**
   * @generated from field: string style_field = 2;
   */
  styleField: string;

  /**
   * @generated from field: string style_value = 3;
   */
  styleValue: string;

  constructor(data?: PartialMessage<DeckStyleUpdate>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.DeckStyleUpdate";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeckStyleUpdate;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeckStyleUpdate;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeckStyleUpdate;

  static equals(a: DeckStyleUpdate | PlainMessage<DeckStyleUpdate> | undefined, b: DeckStyleUpdate | PlainMessage<DeckStyleUpdate> | undefined): boolean;
}

/**
 * @generated from message trackstar.ConfigDeckStyle
 */
export declare class ConfigDeckStyle extends Message<ConfigDeckStyle> {
  /**
   * @generated from field: map<string, string> properties = 1;
   */
  properties: { [key: string]: string };

  constructor(data?: PartialMessage<ConfigDeckStyle>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.ConfigDeckStyle";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigDeckStyle;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigDeckStyle;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigDeckStyle;

  static equals(a: ConfigDeckStyle | PlainMessage<ConfigDeckStyle> | undefined, b: ConfigDeckStyle | PlainMessage<ConfigDeckStyle> | undefined): boolean;
}

/**
 * @generated from message trackstar.Config
 */
export declare class Config extends Message<Config> {
  /**
   * @generated from field: repeated string deck_ids = 1;
   */
  deckIds: string[];

  /**
   * @generated from field: map<string, trackstar.ConfigDeckStyle> style = 2;
   */
  style: { [key: string]: ConfigDeckStyle };

  constructor(data?: PartialMessage<Config>);

  static readonly runtime: typeof proto3;
  static readonly typeName = "trackstar.Config";
  static readonly fields: FieldList;

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Config;

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Config;

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Config;

  static equals(a: Config | PlainMessage<Config> | undefined, b: Config | PlainMessage<Config> | undefined): boolean;
}
