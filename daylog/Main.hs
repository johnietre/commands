import Data.Char (toUpper)
import Data.List (elemIndex)
import Data.Maybe (fromMaybe)
import Data.Time.Clock (addUTCTime, UTCTime)
import Data.Time.Clock.POSIX (posixSecondsToUTCTime, POSIXTime)
import Data.Time.Format (defaultTimeLocale, formatTime, parseTimeM)
import Data.Time.LocalTime (getCurrentTimeZone, localTimeToUTC, utcToLocalTime, LocalTime, TimeZone)
import System.Exit (die)
import System.Environment (getArgs, lookupEnv)
import System.FilePath (joinPath, FilePath)
import System.IO (appendFile, hFlush, readFile, stdout, writeFile, FilePath)
import System.Posix.Time (epochTime)
import System.Posix.Types (EpochTime)

main = do
  mbase <- lookupEnv "JCMDS_PATH"
  base <- case mbase of
    Just base -> return base
    Nothing -> die "\"JCMDS_PATH\" environment variable not set"
  let path = joinPath [base,"daylog","daylog.log"]

  args <- getArgs
  case args of
    [] -> printHelp
    ["add"] -> addLog path
    ("add":"at":time) -> die "\"add at [time]\" command not implemented"
    ("add":"on":day) -> die "\"add on [day]\" command not implemented"
    ["read","all"] -> readAllLogs path
    ["read","last"] -> readLastLog path
    ["read","from",start] -> readLogsFrom path start
    ["read","until",end] -> readLogsUntil path end
    ["read","between",start,end] -> readLogsBetween path start end
    ["read","on",day] -> readLogsOn path day
    ["clear"] -> clearLogs path
    ["help"] -> printHelp
    args -> die $ "Unknown command: " ++ (unwords args)

addLog :: FilePath -> IO ()
addLog path = do
  t <- epochTime
  let s = show t
  tz <- getCurrentTimeZone
  putStrLn $ "Current Time: " ++ (stringToFormat tz s)
  putStrLn $ "What do you want to log?"
  log <- getLine
  appendFile path $ s ++ "|" ++ log ++ "\n"

addLogAt :: FilePath -> String -> IO ()
addLogAt path timeStr = do
  t <- case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" timeStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  return ()

addLogOn :: FilePath -> String -> IO ()
addLogOn path dayStr = do
  t <- case parseTimeM True defaultTimeLocale "%m/%d/%Y" dayStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  return ()

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

readLogsFrom :: FilePath -> String -> IO ()
readLogsFrom path startStr = do
  start <- case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" startStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  let utcTimes = map tupToUTC . map tupToPOSIX . map tupToEpoch . mapSplit $ logLines
      localTimes = map (tupToLocal tz) utcTimes
      filteredLocalTimes = dropWhile (\(t, _) -> t < start) localTimes
  mapM_ putStrLn . map tupToString . map tupToFormat $ filteredLocalTimes

readLogsUntil :: FilePath -> String -> IO ()
readLogsUntil path endStr = do
  end <- case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" endStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  let utcTimes = map tupToUTC . map tupToPOSIX . map tupToEpoch . mapSplit $ logLines
      localTimes = map (tupToLocal tz) utcTimes
      filteredLocalTimes = takeWhile (\(t, _) -> t <= end) localTimes
  mapM_ putStrLn . map tupToString . map tupToFormat $ filteredLocalTimes

readLogsBetween :: FilePath -> String -> String -> IO ()
readLogsBetween path startStr endStr = do
  start <- case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" startStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  end <- case parseTimeM True defaultTimeLocale "%H:%M %m/%d/%Y" endStr :: Maybe LocalTime of
    Just t -> return t
    Nothing -> die "Invalid time format, expects HH:MM mm/dd/YYYY"
  logLines <- readLogLines path
  tz <- getCurrentTimeZone
  let utcTimes = map tupToUTC . map tupToPOSIX . map tupToEpoch . mapSplit $ logLines
      localTimes = map (tupToLocal tz) utcTimes
      timesAfterStart = dropWhile (\(t, _) -> t < start) localTimes
      filteredLocalTimes = takeWhile (\(t, _) -> t <= end) timesAfterStart
  mapM_ putStrLn . map tupToString . map tupToFormat $ filteredLocalTimes

readLogsOn :: FilePath -> String -> IO ()
readLogsOn path dayStr = do
  tz <- getCurrentTimeZone
  sut <- case parseTimeM True defaultTimeLocale "%m/%d/%Y" dayStr :: Maybe LocalTime of
    Just t -> return $ localTimeToUTC tz t
    Nothing -> die "Invalid time format, expects mm/dd/YYYY"
  logLines <- readLogLines path
  let start = utcToLocal tz sut
      end = utcToLocal tz . addUTCTime (60*60*24) $ sut
      utcTimes = map tupToUTC . map tupToPOSIX . map tupToEpoch . mapSplit $ logLines
      localTimes = map (tupToLocal tz) utcTimes
      timesAfterStart = dropWhile (\(t, _) -> t < start) localTimes
      filteredLocalTimes = takeWhile (\(t, _) -> t < end) timesAfterStart
  mapM_ putStrLn . map tupToString . map tupToFormat $ filteredLocalTimes

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
  putStrLn "    add at [time]\t\tAdd log at the given time"
  putStrLn "    add on [day]\t\tAdd log on the given day"
  putStrLn "    read all\t\t\tRead all logs"
  putStrLn "    read last\t\t\tRead the last log"
  putStrLn "    read from [start]\t\tRead logs at or after the start time"
  putStrLn "    read until [end]\t\tRead logs before or at the end time"
  putStrLn "    read between [start] [end]\tRead logs between the start and end times (inclusive)"
  putStrLn "    read on [day]\t\tRead logs for the given day"
  putStrLn "    clear\t\t\tClear all logs"
  putStrLn "    help\t\t\tprint help (this screen)"
  putStrLn "All times must be one argument with the format \"HH:MM mm/dd/YYYY\""
  putStrLn "    Example: 14:45 04/09/2007"
  putStrLn "    Example: 08:00 12/15/2021"
  putStrLn "All days must be one argument with the format \"mm/dd/YYYY\""

-- |The 'readLogLines' function reads the logs from the log file and returns
-- them as a list of strings (lines)
readLogLines :: FilePath -> IO [String]
readLogLines path = do
  contents <- readFile path
  return $ lines contents

-- |The 'splitOnce' function takes a character and string and splits the string
-- over that character once, returning a tuple of the two parts (without the
-- character). If the character is not present, the enter string is the first
-- tuple element
splitOnce :: Char -> String -> (String, String)
splitOnce p s = (take i s, drop (i+1) s)
  where i = fromMaybe (length s) $ p `elemIndex` s

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

-- |The 'tupStringToString' takes a tuple with an epoch as a string as the first
-- element and a timezone, applies the 'tupStringToFormat' function to it, and
-- converts the resulting tuple into a single string
tupStringToString :: TimeZone -> (String, String) -> String
tupStringToString tz = tupToString . tupStringToFormat tz 

-- |The 'mapSplit' function takes a list of strings and maps it with the
-- 'splitOnce' function with the split character as '|'
mapSplit :: [String] -> [(String, String)]
mapSplit = map (splitOnce '|')
