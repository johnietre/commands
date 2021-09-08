#ifndef CLOGGER_H
#define CLOGGER_H

typedef enum {
  LOG_DEBUG,
  LOG_INFO,
  LOG_WARNING,
  LOG_ERROR,
  LOG_CRITICAL,
} LOG_LEVEL;

int set_log_file(const char *filename);
int set_log_level(LOG_LEVEL level);
void log_debug(const char *msg);
void log_info(const char *msg);
void log_warning(const char *msg);
void log_error(const char *msg);
void log_critical(const char *msg);

#endif
