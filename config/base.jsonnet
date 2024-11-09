local hours(n) = n * 60 * 60;
local minutes(n) = n * 60;

local params(top, env, service, params) = 
  { [top]: {
    [s[1]]: s[2][env], 
    for s in params if s[0] == service 
    },
  };

{
  hours:: hours,
  minutes:: minutes,
  params:: params,
}