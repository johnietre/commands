import qualified Data.Char as DC
import System.Environment (getArgs, getProgName)
import System.Exit (die)

main :: IO ()
main = do
  (expr : _) <- getArgs
  if null expr
    then die "Usage: math [expression]"
    else check expr

check :: String -> IO ()
check expr = iterativeCheck expr '\n' 0 False False False
  where
    -- iterativeCheck parameters: expr prev parens openParens decimal two
    iterativeCheck :: String -> Char -> Int -> Bool -> Bool -> Bool -> IO ()
    iterativeCheck "" _ 0 _ _ _ = putStr ""
    iterativeCheck "" _ _ _ _ _ = die "Mismatch parens"
    iterativeCheck ('.' : expr) prev parens openParens decimal _
      | decimal = die "Invalid expression"
      | otherwise = iterativeCheck expr '.' parens openParens True False
    iterativeCheck ('(' : expr) _ parens _ _ _ =
      iterativeCheck expr '(' (parens + 1) True False False
    iterativeCheck (')' : expr) _ parens openParens _ _
      | parens == 0 = die "Mismatch parentheses"
      | otherwise = iterativeCheck expr ')' (parens - 1) (parens /= 1) False False
    iterativeCheck (' ' : expr) prev parens openParens _ two =
      iterativeCheck expr prev parens openParens False two
    iterativeCheck (c : expr) prev parens openParens decimal two
      | DC.isDigit c = iterativeCheck expr c parens openParens decimal two
      | c == '*' || c == '/' =
        if prev == c
          then
            if two
              then die "Invalid expression"
              else iterativeCheck expr c parens openParens False True
          else iterativeCheck expr c parens openParens False False
      | c == '+' || c == '^' || c == '-' || c == '%' =
        if prev == c
          then die "Invalid expression "
          else iterativeCheck expr c parens openParens False two
      | otherwise = die "Invalid character"

-- check :: String -> IO ()
-- check expr = iterativeCheck expr '\n' 0 False False False
--   where
--     iterativeCheck "" _ 0 _ _ _ = putStr ""
--     iterativeCheck "" _ _ _ _ _ = die "Mismatch parens"
--     iterativeCheck (' ' : expr) prev parens openParens decimal two =
--       iterativeCheck expr prev parens openParens decimal two
--     iterativeCheck ('\t' : expr) prev parens openParens decimal two =
--       iterativeCheck expr prev parens openParens decimal two
--     iterativeCheck (c : expr) prev parens openParens decimal two
--       | DC.isDigit c = iterativeCheck expr c parens openParens decimal two
--       | c == '.' =
--         if decimal
--           then die "Invalid expression"
--           else iterativeCheck expr c parens openParens True False
--       | c == '*' || c == '/' =
--         if prev == c
--           then
--             if two
--               then die "Invalid expression"
--               else iterativeCheck expr c parens openParens False True
--           else iterativeCheck expr c parens openParens False False
--       | c == '+' || c == '^' || c == '-' || c == '%' =
--         if prev == c
--           then die "Invalid expression "
--           else iterativeCheck expr c parens openParens False two
--       | c == '(' =
--         if not openParens
--           then iterativeCheck expr c (parens + 1) True False two
--           else iterativeCheck expr c (parens + 1) openParens False two
--       | c == ')' =
--         if parens == 0
--           then die "Mismatch parentheses"
--           else
--             if parens == 1
--               then iterativeCheck expr c (parens - 1) False False two
--               else iterativeCheck expr c (parens - 1) openParens False two
--       | otherwise = die "Invalid character"

calc :: Double -> String -> Double -> Double
calc x "+" y = x + y
calc x "-" y = x - y
calc x "*" y = x * y
calc x "/" y = x / y
calc x "^" y = x ** y
calc x "**" y = x ** y
calc x "//" y = fromIntegral (floor (x / y))
-- Making sure inputs are integral is be job of caller
calc x "%" y = fromIntegral (round x `mod` round y)
calc x "log" y = logBase x y
calc x "ln" _ = log x
calc x "sin" _ = sin x
calc x "cos" _ = cos x
calc x "tan" _ = tan x
-- Making sure inputs are integral is be job of caller
calc x "!" _ = factorial (round x)
calc _ _ _ = 0 / 0

factorial :: Int -> Double
factorial 1 = 1
factorial x = fromIntegral x * factorial (x - 1)

factorialInteger :: Integer -> Integer
factorialInteger 1 = 1
factorialInteger x = x * factorialInteger (x - 1)