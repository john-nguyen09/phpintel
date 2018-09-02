import { Union, Literal, Static } from "runtypes";

export const IndexNames = Union(
    Literal('IDENTIFIER'),
    Literal('POSITION'),
    Literal('TIMESTAMP'),
    Literal('URI')
);


export type IndexNames = Static<typeof IndexNames>;

export const IndexId: {[key in IndexNames]: string} = {
    IDENTIFIER: 'identifier_index',
    POSITION: 'position_index',
    TIMESTAMP: 'timestamp_index',
    URI: 'uri_index',
}

export const IndexVersion: {[key in IndexNames]: number} = {
    IDENTIFIER: 1,
    POSITION: 1,
    TIMESTAMP: 1,
    URI: 1
}