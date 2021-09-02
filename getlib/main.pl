#!/usr/bin/perl
use strict;

my $libspath = "/home/johnierodgers/.storage/libs";

my $argc = @ARGV;
if ($ARGV[0] eq "-l") {
  opendir(DIR, $libspath) or die "Error opening libs folder, $!";
  my $file;
  while ($file = readdir DIR) {
    if ($file ne "." and $file ne "..") {
      print "$file\n";
    }
  }
  closedir DIR;
  exit;
} elsif ($ARGV[0] eq "cpp-server") {
  `cp -r $libspath/cpp-server .`;
  exit;
} elsif ($argc != 2) {
  die "Usage: libs {libname} {lang_ext}\n";
}

my $libname = $ARGV[0];
my $ext = $ARGV[1];
if ($ext eq "c" or $ext eq "cpp") {
  $ext =~ s/c/h/;
}
`cp $libspath/$libname/$libname.$ext .`
