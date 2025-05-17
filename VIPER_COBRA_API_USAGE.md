# Comprehensive List of Viper and Cobra API Usage

## Cobra API Calls and Locations

### Command Creation and Structure
- `&cobra.Command{}` - main.go:34 (rootCmd), and for each subcommand
- `rootCmd.AddCommand()` - main.go:67-71 (adding subcommands)
- `cmd.Use` - Used in all command definitions
- `cmd.Short` - Used in all command definitions
- `cmd.Long` - Used in rootCmd and completion command
- `cmd.Version` - main.go:40
- `cmd.Args` - main.go:636, 723, 882, 1064
- `cmd.Run` - Used in all command definitions
- `cmd.ValidArgs` - main.go:1064
- `cmd.DisableFlagsInUseLine` - main.go:1062

### Flag Management
- `cmd.PersistentFlags().StringVar()` - main.go:47
- `cmd.PersistentFlags().String()` - main.go:48-49
- `cmd.PersistentFlags().BoolP()` - main.go:50
- `cmd.PersistentFlags().StringP()` - main.go:51
- `cmd.PersistentFlags().Lookup()` - main.go:53, 56, 59, 62
- `cmd.Flags().StringP()` - main.go:621, 871, 873, 1015, 1016
- `cmd.Flags().Float32P()` - main.go:622, 712, 874, 1018
- `cmd.Flags().Int()` - main.go:623, 713, 1019
- `cmd.Flags().Bool()` - main.go:624, 714, 1017
- `cmd.Flags().StringSliceP()` - main.go:871
- `cmd.Flags().GetString()` - main.go:510, 738, 741, 873, 899, 904
- `cmd.Flags().GetFloat32()` - main.go:511, 652, 744
- `cmd.Flags().GetInt()` - main.go:512, 653, 922
- `cmd.Flags().GetBool()` - main.go:513, 514, 654, 742, 924
- `cmd.Flags().GetStringSlice()` - main.go:739
- `cmd.Flags().Changed()` - main.go:521
- `cmd.MarkFlagsMutuallyExclusive()` - main.go:628

### Command Validation
- `cobra.MinimumNArgs()` - main.go:636, 723, 882
- `cobra.ExactArgs()` - main.go:1064
- `cobra.OnlyValidArgs` - main.go:1064
- `cobra.MatchAll()` - main.go:1064

### Shell Completion
- `cmd.Root().GenBashCompletion()` - main.go:1068
- `cmd.Root().GenZshCompletion()` - main.go:1070
- `cmd.Root().GenFishCompletion()` - main.go:1072
- `cmd.Root().GenPowerShellCompletionWithDesc()` - main.go:1074

### Initialization and Execution
- `cobra.OnInitialize()` - main.go:45
- `cobra.CheckErr()` - main.go:79
- `rootCmd.Execute()` - main.go:105

## Viper API Calls and Locations

### Configuration File Management
- `viper.SetConfigFile()` - main.go:76
- `viper.AddConfigPath()` - main.go:81, 82, 88
- `viper.SetConfigType()` - main.go:83
- `viper.SetConfigName()` - main.go:84, 89
- `viper.ReadInConfig()` - main.go:99
- `viper.ConfigFileUsed()` - main.go:100

### Environment Variables
- `viper.AutomaticEnv()` - main.go:92

### Default Values
- `viper.SetDefault()` - main.go:95-97

### Flag Binding
- `viper.BindPFlag()` - main.go:53, 56, 59, 62

### Getting Values
- `viper.GetString()` - main.go:113, 117, 119, 130, 133

### Testing-specific (main_test.go)
- `viper.Set()` - main_test.go:61, 78
- `viper.Reset()` - main_test.go:19

## File-by-File Breakdown

### `cmd/main.go`

Lines with viper usage:
- 28: import "github.com/spf13/viper"
- 53-64: viper.BindPFlag() calls
- 76: viper.SetConfigFile()
- 81-89: viper config path setup
- 92: viper.AutomaticEnv()
- 95-97: viper.SetDefault()
- 99-100: viper.ReadInConfig()
- 113-119: viper.GetString() in getProvider()
- 130-133: viper.GetString() in getAPIKey()

Lines with cobra usage:
- 27: import "github.com/spf13/cobra"
- 34-41: rootCmd creation
- 45: cobra.OnInitialize()
- 47-51: PersistentFlags setup
- 53-64: flag binding
- 67-71: rootCmd.AddCommand()
- 79: cobra.CheckErr()
- 105: rootCmd.Execute()
- 490-629: newChatCmd() implementation
- 632-716: newCompleteCmd() implementation
- 718-876: newAgentCmd() implementation
- 878-1021: newStructuredCmd() implementation
- 1023-1079: newCompletionCmd() implementation

### `cmd/main_test.go`

Lines with viper usage:
- 7: import "github.com/spf13/viper"
- 19: viper.Reset()
- 61: viper.Set()
- 78: viper.Set()

## Integration Points Summary

1. **Configuration Loading** (initConfig function):
   - Called via cobra.OnInitialize()
   - Uses viper for all config management

2. **Provider Configuration** (getProvider function):
   - Reads from viper for provider and model
   - Falls back to provider-specific defaults

3. **API Key Resolution** (getAPIKey function):
   - First tries viper config
   - Falls back to environment variables

4. **Flag Binding** (init function):
   - Binds all persistent flags to viper
   - Allows config file override from CLI

5. **Command Implementation**:
   - All commands use cobra flag parsing
   - Values retrieved through either cobra flags or viper