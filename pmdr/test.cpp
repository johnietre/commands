#define MINIAUDIO_IMPLEMENTATION
#include "miniaudio.h"
#include <iostream>
#include <string>
using namespace std;

ma_result result;
ma_engine engine;
ma_sound sound;

int main(int argc, char **argv) {
  result = ma_engine_init(NULL, &engine);
  if (result != MA_SUCCESS)
    return result;
  ma_engine_play_sound(&engine, "OdessaUp.wav", NULL);
  getchar();
  ma_engine_uninit(&engine);
  return 0;
}
