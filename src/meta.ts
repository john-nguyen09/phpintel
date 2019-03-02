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

export namespace Position {
    export function contains(start: Position, end: Position, target: Position): boolean {
        if (target.row < start.row || target.row > end.row) {
            return false;
        }

        if (target.row === start.row && target.column < start.column) {
            return false;
        }

        if (target.row === end.row && target.column > end.column) {
            return false;
        }

        return true;
    }
}