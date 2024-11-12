package main

import (
    "fmt"
    "os"
    "os/exec"

    "github.com/pelletier/go-toml"
    "github.com/spf13/cobra"
)

// 構造体の定義
type FontConfig struct {
    Size        float64 `toml:"size,omitempty"`
    Family      string  `toml:"family,omitempty"`
    BoldFamily  string  `toml:"bold,omitempty"`
    ItalicFamily string `toml:"italic,omitempty"`
}

type WindowConfig struct {
    X            int     `toml:"x,omitempty"`
    Y            int     `toml:"y,omitempty"`
    Opacity      float64 `toml:"opacity,omitempty"`
    Columns      int     `toml:"columns,omitempty"`
    Lines        int     `toml:"lines,omitempty"`
    Title        string  `toml:"title,omitempty"`
    StartupMode  string  `toml:"startup_mode,omitempty"`
    Decorations  string  `toml:"decorations,omitempty"`
    DynamicTitle bool    `toml:"dynamic_title,omitempty"`
}

type CursorConfig struct {
    Style string `toml:"style,omitempty"`
    Blink bool   `toml:"blink,omitempty"`
}

type AlacrittyConfig struct {
    Font      *FontConfig      `toml:"font,omitempty"`
    Window    *WindowConfig    `toml:"window,omitempty"`
    Cursor    *CursorConfig    `toml:"cursor,omitempty"`
}

// 候補が限定されているオプションの値リスト
var validCursorStyles = []string{"Block", "Underline", "Beam"}
var validWindowStartupModes = []string{"Windowed", "Maximized", "Fullscreen", "SimpleFullscreen"}
var validWindowDecorations = []string{"Full", "None", "Transparent", "Buttonless"}

func main() {
    var rootCmd = &cobra.Command{Use: "alaconf"}

    // 設定変更用の変数
    var fontSize float64
    var fontFamily, fontBoldFamily, fontItalicFamily string
    var cursorStyle string
    var windowStartupMode, windowDecorations string
		var windowOpacity float64

    var configCmd = &cobra.Command{
        Use:   "config",
        Short: "複数のAlacritty設定を同時に変更します",
        Run: func(cmd *cobra.Command, args []string) {
            config := AlacrittyConfig{}

            // フォント設定のチェックと適用
            if fontSize > 0 || fontFamily != "" || fontBoldFamily != "" || fontItalicFamily != "" {
                config.Font = &FontConfig{
                    Size:         fontSize,
                    Family:       fontFamily,
                    BoldFamily:   fontBoldFamily,
                    ItalicFamily: fontItalicFamily,
                }
            }

            // カーソルスタイルのチェックと適用
            if cursorStyle != "" && !isValidOption(cursorStyle, validCursorStyles) {
                fmt.Printf("エラー: cursor-styleには次のいずれかを指定してください: %v\n", validCursorStyles)
                return
            }
            config.Cursor = &CursorConfig{Style: cursorStyle}

            // ウィンドウ設定のチェックと適用
            if windowStartupMode != "" && !isValidOption(windowStartupMode, validWindowStartupModes) {
                fmt.Printf("エラー: window-startup-modeには次のいずれかを指定してください: %v\n", validWindowStartupModes)
                return
            }
            if windowDecorations != "" && !isValidOption(windowDecorations, validWindowDecorations) {
                fmt.Printf("エラー: window-decorationsには次のいずれかを指定してください: %v\n", validWindowDecorations)
                return
            }
            config.Window = &WindowConfig{
							Opacity: windowOpacity,
                StartupMode:  windowStartupMode,
                Decorations:  windowDecorations,
            }

            // TOML形式に変換してAlacrittyに送信
            tomlData, err := toml.Marshal(config)
            if err != nil {
                fmt.Printf("TOMLへの変換に失敗しました: %v\n", err)
                return
            }

            runAlacrittyConfig(string(tomlData))
        },
    }

    // フォント設定のフラグ
    configCmd.Flags().Float64Var(&fontSize, "font-size", 0, "フォントサイズを指定します")
    configCmd.Flags().StringVar(&fontFamily, "font-family", "", "フォントファミリーを指定します")
    configCmd.Flags().StringVar(&fontBoldFamily, "font-bold-family", "", "太字用のフォントファミリーを指定します")
    configCmd.Flags().StringVar(&fontItalicFamily, "font-italic-family", "", "イタリック用のフォントファミリーを指定します")

    // カーソルとウィンドウのフラグ
    configCmd.Flags().StringVar(&cursorStyle, "cursor-style", "", "カーソルのスタイルを指定します (Block, Underline, Beam)")
    configCmd.Flags().StringVar(&windowStartupMode, "window-startup-mode", "", "ウィンドウの起動モードを指定します (Windowed, Maximized, Fullscreen, SimpleFullscreen)")
    configCmd.Flags().StringVar(&windowDecorations, "window-decorations", "", "ウィンドウ装飾を指定します (Full, None, Transparent, Buttonless)")
		configCmd.Flags().Float64Var(&windowOpacity, "window-opacity", 1.0, "ウィンドウの不透明度を指定します (0.0〜1.0)")

    rootCmd.AddCommand(configCmd)

    // リセットコマンドの追加
    var resetCmd = &cobra.Command{
        Use:   "reset",
        Short: "Alacrittyの設定をリセットします",
        Run: func(cmd *cobra.Command, args []string) {
            runAlacrittyReset()
        },
    }
    rootCmd.AddCommand(resetCmd)

    // 補完の初期化
    initCompletion(rootCmd)

    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

// 補完の初期化
func initCompletion(rootCmd *cobra.Command) {
    rootCmd.RegisterFlagCompletionFunc("cursor-style", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return validCursorStyles, cobra.ShellCompDirectiveNoFileComp
    })
    rootCmd.RegisterFlagCompletionFunc("window-startup-mode", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return validWindowStartupModes, cobra.ShellCompDirectiveNoFileComp
    })
    rootCmd.RegisterFlagCompletionFunc("window-decorations", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return validWindowDecorations, cobra.ShellCompDirectiveNoFileComp
    })
}

// 入力値が有効かチェックするユーティリティ関数
func isValidOption(value string, validOptions []string) bool {
    for _, option := range validOptions {
        if value == option {
            return true
        }
    }
    return false
}

// Alacrittyの設定を動的に変更する関数
func runAlacrittyConfig(configStr string) {
    cmd := exec.Command("alacritty", "msg", "config", configStr)
    if err := cmd.Run(); err != nil {
        fmt.Printf("Alacrittyの設定変更に失敗しました: %v\n", err)
    } else {
        fmt.Println("設定が適用されました:", configStr)
    }
}

// Alacrittyの設定をリセットする関数
func runAlacrittyReset() {
    cmd := exec.Command("alacritty", "msg", "config", "-r")
    if err := cmd.Run(); err != nil {
        fmt.Printf("Alacrittyの設定リセットに失敗しました: %v\n", err)
    } else {
        fmt.Println("Alacrittyの設定がリセットされました")
    }
}

