local util = import 'domain64/util.libsonnet';

{
  all: [
    0,
    18446744073709551616,
  ],
  three: [
    [
      72057748656750592,
      72057748656750592,
    ],
    [
      72058358542106624,
      72058358542106624,
    ],
    [
      72057911865508096,
      72057911865508096,
    ],
    [
      72058358542106624,
      72058362837073664,
    ],
  ],
  nih: [
    util.toDec('0100008D00000000'),
    util.toDec('0100008DFFFFFF00'),
  ],
  uscg: [
    util.toDec('0300002000000000'),
    util.toDec('0300002FF0000000'),
  ],
  spaceforce: [
    util.toDec('0300001E00000000'),
    util.toDec('0300001EFF000000'),
  ],
  nasa: [
   util.toDec("0100008700000000"),
   util.toDec("01000087FF000000")
  ],
  dec15: [
    self.nih + self.three + self.uscg + self.spaceforce, self.nasa
  ],
}
