# dp(1) - Directory Pipe

    dp(1)                       General Commands Manual                      dp(1)
    
    NAME
           dp — directory pipe 1.0
    
    SYNOPSIS
           dp [-h]
           dp [-k] [-i '%i'] [-o '%o'] [-t '/tmp'] cmd [args ...]
    
    DESCRIPTION
           dp  wraps the execution of a command working from an input directory to
           an output directory. The main goal is to chain such  wrapped  commands,
           so  as  to progressively compile an input directory to an output direc‐
           tory, in a series of simple passes, without having to deal with  creat‐
           ing  and  removing  temporary locations. The chaining is expected to be
           performed by  reading/writing  tar(1)  compressed  directories  through
           standard UNIX pipe(7).
    
           This  can  be  useful for prototyping, and more generally, for software
           without strong performance requirements. A static site generator  is  a
           typical use-case.
    
    OPTIONS
           If  at  least  one  of  the  specified  command cmd 's arguments, args,
           matches exactly the placeholder specified by -i, which default  to  %i,
           then dp reads a tar(1) archive from stdin,  uncompresses it to a tempo‐
           rary  location,  systematically replaces the placeholder by the path to
           this temporary location, then finally executes the command.
    
           If no such pattern is specified, dp forwards its stdin to  the  command
           and executes it.
    
           If  at  least  one  of  the  specified  command cmd 's arguments, args,
           matches exactly the placeholder specified by -o, which default  to  %o,
           then  dp systematically replaces the placeholder by the path to a newly
           created temporary location, then executes the command.  After the  com‐
           mand's  execution, this location is compressed to a tar(1) archive, it‐
           self finally sent to stdout.
    
           dp always redirects the command's stdout/stderr to dp   's  stderr.  In
           particular, this avoids messing up an eventual tar(1) output.
    
           The  -t  flag changes the base directory for temporary locations, which
           (should) default to /tmp.
    
           If the command execution fails (non-zero status), no tar(1) archive  is
           sent to stdout, and dp will borrow the command's exit code.
    
           The  temporary  locations are systematically removed unless the -k flag
           is enabled. In which case they are kept, and displayed on stderr before
           processing the input/command.
    
    EXAMPLE
           Here's how an arbitrary three passes directory compilation can  be  im‐
           plemented:
    
                       dp prepare   "$ind" '%o'   \
                     | dp act       '%i'   '%o'   \
                     | dp finish    '%i'   "$outd"
    
    dp 1.0                            $Mdocdate$                             dp(1)
