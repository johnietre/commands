#include "clogger.h"
#include <stdio.h>
#include <string.h>

static FILE *log_file;
static LOG_LEVEL log_level;

int set_log_file(const char *filename) {
  if (filename == NULL)
    return 1;
  if (!strlen(filename)) {
    log_file = stdout;
    return 0;
  }
  return 0;
}

int set_log_level(LOG_LEVEL level) {
  log_level = level;
  return 0;
}

void log_debug(const char *msg) {
  if (log_level > LOG_DEBUG || msg == NULL)
    return;
  fputs(log_file, msg);
}

void log_info(const char *msg) {
  if (log_level > LOG_INFO)
    return;
  fputs(log_file, msg);
}

void log_warning(const char *msg) {
  if (log_level > LOG_WARNING || msg == NULL)
    return;
  fputs(log_file, msg);
}

void log_error(const char *msg) {
  if (log_level > LOG_ERROR || msg == NULL)
    return;
  fputs(log_file, msg);
}

void log_critical(const char *msg) {
  if (msg == NULL)
    return;
  fputs(log_file, msg);
}
