-- TODO: add and delete commands
import           Data.Char
import           Data.List
import           Data.Maybe
import           Data.Time.Clock
import           Data.Time.Clock.POSIX
import           Data.Time.Format
import           Data.Time.LocalTime
import           System.Directory
import           System.Environment
import           System.Exit
import           System.FilePath
import           System.IO
import           System.Posix.Time
import           System.Posix.Types
import           Text.Read

type TimeString = String
type DayString = String

version :: IO ()
version = putStrLn "versionless"

main :: IO ()
main = do
  mbase <- lookupEnv "JCMDS_PATH"
  base <- case mbase of
    Just base -> return base
    Nothing   -> die "\"JCMDS_PATH\" environment variable not set"
  let path = joinPath [base,"daylog","daylog.log"]

  args <- getArgs
  case args of
    [] -> printHelp
    ["add"] -> addLog path
    ["add","at",time] -> addLogAt path time
    ["add","on",day] -> addLogOn path day
    ["read","all"] -> readAllLogs path
    ["read","last"] -> readLastLog path
    ("read":"tail":num) -> readTailLogs path $ intercalate " " num
    ["read","from",start] -> readLogsFrom path start
    ["read","until",end] -> readLogsUntil path end
    ["read","between",start,end] -> readLogsBetween path start end
    ["read","on",day] -> readLogsOn path day
    ["delete"] -> deleteLog path
    ["delete","from",start] -> deleteLogsFrom path start
    ["delete","until",end] -> deleteLogsUntil path end
    ["delete","between",start,end] -> deleteLogsBetween path start end
    ["delete","at",time] -> deleteLogsAt path time
    ["delete","on",day] -> deleteLogOn path day
    ["clear"] -> clearLogs path
    ["version"] -> version
    ["help"] -> printHelp
    _ -> die $ "Unknown command: " ++ (unwords args)

addLog :: FilePath -> IO ()
addLog path = do
  t <- epochTime
  let s = show t
  tz <- getCurrentTimeZone
  putStrLn $ "Current Time: " ++ (stringToFormat tz s)
  putStrLn $ "What do you want to log?"
  logMsg <- getLine
  if null logMsg
    then return ()
    else appendFile path $ s ++ "|" ++ logMsg ++ "\n"

addLogAt :: FilePath -> TimeString -> IO ()
addLogAt path timeStr = die "\"add at [time]\" command not implemented"

addLogOn :: FilePath -> DayString -> IO ()
addLogOn path dayStr = die "\"add on [day]\" command not implemented"

readAllLogs :: FilePath -> IO ()
readAllLogs path = do
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  mapM_ putStrLn . map (tupStringToString tz) . mapSplit $ logLines

readLastLog :: FilePath -> IO ()
readLastLog path = do
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  putStrLn . last . map (tupStringToString tz) . mapSplit $ logLines

readTailLogs :: FilePath -> String -> IO ()
readTailLogs path numStr = do
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  let l = length logLines
  num <- if null numStr
            then return $ l-1
            else case readMaybe numStr of
                  Just n  -> return $ if n > 0 then n else l-1
                  Nothing -> die "Invalid number"
  mapM_ putStrLn
    . drop (l-num)
    . map (tupStringToString tz)
    . mapSplit $ logLines

readLogsFrom :: FilePath -> TimeString -> IO ()
readLogsFrom path startStr = do
  start <- parseLocalTime startStr
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  mapM_ putStrLn
    . map tupToString
    . map tupToFormat
    . tupFrom start
    . map (tupToLocal tz)
    . map tupToUTC
    . map tupToPOSIX
    . map tupToEpoch
    . mapSplit $ logLines

readLogsUntil :: FilePath -> TimeString -> IO ()
readLogsUntil path endStr = do
  end <- parseLocalTime endStr
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  mapM_ putStrLn
    . map tupToString
    . map tupToFormat
    . tupUntil end
    . map (tupToLocal tz)
    . map tupToUTC
    . map tupToPOSIX
    . map tupToEpoch
    . mapSplit $ logLines

readLogsBetween :: FilePath -> TimeString -> TimeString -> IO ()
readLogsBetween path startStr endStr = do
  start <- parseLocalTime startStr
  end <- parseLocalTime endStr
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  mapM_ putStrLn
    . map tupToString
    . map tupToFormat
    . tupBetween start end
    . map (tupToLocal tz)
    . map tupToUTC
    . map tupToPOSIX
    . map tupToEpoch
    . mapSplit $ logLines

readLogsOn :: FilePath -> DayString -> IO ()
readLogsOn path dayStr = do
  tz <- getCurrentTimeZone
  s <- parseLocalDay dayStr
  logLines <- readLogLines path
  let sut = localTimeToUTC tz s
      start = utcToLocal tz sut
      end = utcToLocal tz . addUTCTime (60*60*24-1) $ sut
   in mapM_ putStrLn
      . map tupToString
      . map tupToFormat
      . tupBetween start end
      . map (tupToLocal tz)
      . map tupToUTC
      . map tupToPOSIX
      . map tupToEpoch
      . mapSplit $ logLines

deleteLog :: FilePath -> IO ()
deleteLog path = do
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  mapM_ putStrLn
    . map (\(i, s) -> show (i+1) ++ ") " ++ s)
    . enumerate
    . map (tupStringToString tz)
    . mapSplit $ logLines
  putStr "Choice: "
  hFlush stdout
  numStr <- getLine
  num <- case readMaybe numStr of
    Just n -> if n < 1 || n > length logLines
                 then die "Invalid choice"
                 else return n
    Nothing -> die "Invalid choice"
  (tempName, tempHandle) <- openTempFile (dropFileName path) "temp"
  hPutStr tempHandle . unlines . delete (logLines !! (num-1)) $ logLines
  hClose tempHandle
  removeFile path
  renameFile tempName path

deleteLogsFrom :: FilePath -> String -> IO ()
deleteLogsFrom path timeStr = die "delete"

deleteLogsUntil :: FilePath -> String -> IO ()
deleteLogsUntil path timeStr = die "delete"

deleteLogsBetween :: FilePath -> String -> String -> IO ()
deleteLogsBetween path startStr endStr = die "delete"

deleteLogsAt :: FilePath -> TimeString -> IO ()
deleteLogsAt path timeStr = die "\"delete at [time]\" command not implemented"

deleteLogOn :: FilePath -> DayString -> IO ()
deleteLogOn path dayStr = die "\"delete on [day]\" command not implemented"

clearLogs :: FilePath -> IO ()
clearLogs path = do
  putStr "Clear logs? [Y/n] "
  hFlush stdout
  c <- getChar
  if toUpper c == 'Y' then writeFile path "" else return ()

printHelp :: IO ()
printHelp = do
  putStrLn "Daylog is used to log thigs throughout the day"
  putStrLn "\t\tUsage: daylog <command>"
  putStrLn "    add\t\t\t\tAdd a new log with the current time"
  putStrLn "\tat [time]\t\tAdd log at the given time"
  putStrLn "\ton [day]\t\tAdd log on the given day"
  putStrLn "    read"
  putStrLn "\tall\t\t\tRead all logs"
  putStrLn "\tlast\t\t\tRead the last log"
  putStrLn "\ttail [num]\t\tRead the last n logs if given, otherwise all but the first"
  putStrLn "\tfrom [start]\t\tRead logs at or after the start time"
  putStrLn "\tuntil [end]\t\tRead logs before or at the end time"
  putStrLn "\tbetween [start] [end]\tRead logs between the start and end times (inclusive)"
  putStrLn "\ton [day]\t\tRead logs for the given day"
  putStrLn "    delete\t\t\tDelete a log from all logs"
  putStrLn "\tfrom [start]\t\tDelete logs at or after the start time"
  putStrLn "\tuntil [end]\t\tDelete logs before or at the end time"
  putStrLn "\tbetween [start] [end]\tDelete logs between the start and end times (inclusive)"
  putStrLn "\tat [time]\t\tDelete the log(s) at the given time"
  putStrLn "\ton [day]\t\tDelete a log on the given day"
  putStrLn "    clear\t\t\tClear all logs"
  putStrLn "    version\t\t\tPrint daylog version"
  putStrLn "    help\t\t\tPrint help (this screen)"
  putStrLn ""
  putStrLn "All times must be one argument with the format \"HH:MM mm/dd/YYYY\""
  putStrLn "    Example: 14:45 04/09/2007"
  putStrLn "    Example: 08:00 12/15/2021"
  putStrLn "All days must be one argument with the format \"mm/dd/YYYY\""
  putStrLn "\"today\" can added as a time or day to use today's date at a time of \"00:00\""

-- |The 'readLogLines' function reads the logs from the log file and returns
-- them as a list of strings (lines)
readLogLines :: FilePath -> IO [String]
readLogLines path = do
  contents <- readFile path
  if null contents then die "No logs" else return $ lines contents

parseLocalDay :: DayString -> IO LocalTime
parseLocalDay "today" = do
  currUTC <- getCurrentTime
  tz <- getCurrentTimeZone
  let local = utcToLocalTime tz currUTC
  return $ local{localTimeOfDay=midnight}
parseLocalDay dayStr = case parseTimeM True defaultTimeLocale "%m/%d/%Y" dayStr :: Maybe LocalTime of
  Just t  -> return t
  Nothing -> die "Invalid time format, expects mm/dd/YYYY"

parseLocalTime :: TimeString -> IO LocalTime
parseLocalTime "today" = parseLocalDay "today"
parseLocalTime timeStr = case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" timeStr :: Maybe LocalTime of
  Just t  -> return t
  Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"

-- |The 'splitOnce' function takes a character and string and splits the string
-- over that character once, returning a tuple of the two parts (without the
-- character). If the character is not present, the enter string is the first
-- tuple element
splitOnce :: Char -> String -> (String, String)
splitOnce p s = (take i s, drop (i+1) s)
  where i = fromMaybe (length s) $ p `elemIndex` s

-- |The 'enumerate' function takes a list and enumerates it
enumerate :: [a] -> [(Int, a)]
enumerate = zip [0..]

-- |The 'stringToEpoch' function takes an epoch as a string and returns it as
-- an 'EpochTime'
stringToEpoch :: String -> EpochTime
stringToEpoch = read

-- |The 'epochToPOSIX' function takes an 'EpochTime' and converts it to a
-- 'POSIXTime'
epochToPOSIX :: EpochTime -> POSIXTime
epochToPOSIX = realToFrac

-- |The 'posixToUTC' function takes a 'POSIXTime' and converts it to a
-- 'UTCTime'
posixToUTC :: POSIXTime -> UTCTime
posixToUTC = posixSecondsToUTCTime

-- |The 'utcToLocal' function takes a 'UTCTime' and timezone and converts it to
-- a 'LocalTime'
utcToLocal :: TimeZone -> UTCTime -> LocalTime
utcToLocal = utcToLocalTime

-- |The 'localToFormat' function takes a 'LocalTime' and converts it to a
-- formatted time string
localToFormat :: LocalTime -> String
localToFormat = formatTime defaultTimeLocale "%H:%M %a, %b %d %Y"
-- localToFormat = formatTime defaultTimeLocale "%H:%M %D"

-- |The 'stringToFormat' function takes an epoch as a string and a timezone
-- and converts it to a formatted time string
stringToFormat :: TimeZone -> String -> String
stringToFormat tz = localToFormat . utcToLocal tz . posixToUTC . epochToPOSIX . stringToEpoch

-- |The 'tupToEpoch' function applies the 'stringToEpoch' function to the first
-- tuple element
tupToEpoch :: (String, String) -> (EpochTime, String)
tupToEpoch (t, s) = (stringToEpoch t, s)

-- |The 'tupToPOSIX' function applies the 'epochToPOSIX' function to the first
-- tuple element
tupToPOSIX :: (EpochTime, String) -> (POSIXTime, String)
tupToPOSIX (e, s) = (epochToPOSIX e, s)

-- |The 'tupToUTC' function applies the 'posixToUTC' function to the first
-- tuple element
tupToUTC :: (POSIXTime, String) -> (UTCTime, String)
tupToUTC (p, s) = (posixToUTC p, s)

-- |The 'tupToLocal' function applies the 'utcToLocal' function to the first
-- tuple element, given a timezone
tupToLocal :: TimeZone -> (UTCTime, String) -> (LocalTime, String)
tupToLocal tz (p, s) = (utcToLocal tz p, s)

-- |The 'tupToFormat' function applies the 'localToFormat' function to the
-- first tuple element
tupToFormat :: (LocalTime, String) -> (String, String)
tupToFormat (l, s) = (localToFormat l, s)

-- |The 'tupToString' function takes a tuple with a formatted time string as
-- the first element and converts the tuple to a string
tupToString :: (String, String) -> String
tupToString (t, s) = t ++ " => " ++ s

-- The 'tupStringToFormat' function applies the 'stringToFormat' function to
-- the first tuple element
tupStringToFormat :: TimeZone -> (String, String) -> (String, String)
tupStringToFormat tz (t, s) = (stringToFormat tz t, s)

-- |The 'tupStringToString' function takes a tuple with an epoch as a string as
-- the first element and a timezone, applies the 'tupStringToFormat' function
-- to it, and converts the resulting tuple into a single string
tupStringToString :: TimeZone -> (String, String) -> String
tupStringToString tz = tupToString . tupStringToFormat tz

-- |The 'mapSplit' function takes a list of strings and maps it with the
-- 'splitOnce' function with the split character as '|'
mapSplit :: [String] -> [(String, String)]
mapSplit = map (splitOnce '|')

-- |The 'tupFrom' function takes a starting time and list of tuples of local
-- times and strings and returns all elements with a time after the start
tupFrom :: LocalTime -> [(LocalTime, String)] -> [(LocalTime, String)]
tupFrom start = dropWhile ((< start) . fst)

-- |The 'tupUntil' function takes a ending time and list of tuples of local
-- times and strings and takes all elements with a time before the end
tupUntil :: LocalTime -> [(LocalTime, String)] -> [(LocalTime, String)]
tupUntil end = takeWhile ((<= end) . fst)

-- |The 'tupBetween' function takes starting and ending times, as well as a
-- list of local times and strings and combines the 'tupFrom' and 'tupUntil'
-- functions
tupBetween :: LocalTime -> LocalTime -> [(LocalTime, String)] -> [(LocalTime, String)]
tupBetween start end = tupUntil end . tupFrom start
