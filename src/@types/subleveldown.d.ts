// Type definitions for level
// Project: https://github.com/Level/subleveldown

declare module 'subleveldown' {
    interface SubLevelDownOptions {
        separator?: string;
        keyEncoding?: Level.Encoding | string;
        valueEncoding?: Level.Encoding | string;
    }

    function subleveldown(
        level: Level.LevelUp,
        prefix: string,
        options?: SubLevelDownOptions
    ): Level.LevelUp;

    export = subleveldown;
}