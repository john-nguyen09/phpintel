export interface Location {
    uri: string;
    range: Range;
}

export interface Position {
    row: number;
    column: number;
}

export interface Range {
    start: Position;
    end: Position;
}