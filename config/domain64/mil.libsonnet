local assertion = import 'assertions.libsonnet';

local domains = {
  '000001': {
    children: {
      '000001': 'aetc.www',
      '000002': 'nellis.www',
      '000003': 'yokota.www',
      '000004': 'e-publishing.www',
      '000005': 'data.www',
      '000006': 'ang.181iw.www',
      '000007': 'netcents.www',
      '000008': 'trademark.www',
      '000009': 'jagreporter.www',
      '00000A': 'ang.118wg.www',
      '00000B': 'ang.189aw.www',
      '00000C': 'ang.155arw.www',
      '00000D': 'ang.138fw.www',
      '00000E': 'afotec.www',
      '00000F': 'ang.186arw.www',
      '000010': 'ang.190arw.www',
      '000011': 'ang.156wg.www',
      '000012': 'ang.113wg.www',
      '000013': 'afhistory.www',
      '000014': 'ang.149fw.www',
      '000015': 'ang.141arw.www',
      '000016': 'ang.187fw.www',
      '000017': 'ang.152aw.www',
      '000018': 'ang.130aw.www',
      '000019': 'ang.183wg.www',
      '00001A': 'ang.126arw.www',
      '00001B': 'veterans-in-blue.www',
      '00001C': 'ang.151arw.www',
      '00001D': '12ftw.www',
      '00001E': 'ang.120thairliftwing.www',
      '00001F': 'ang.158fw.www',
      '000020': 'hq.safie.www',
      '000021': 'ang.110wg.www',
      '000022': 'ang.eads.www',
      '000023': 'ang.153aw.www',
      '000024': 'ang.125fw.www',
      '000025': 'ang.171arw.www',
      '000026': 'ang.124thfighterwing.www',
      '000027': 'ang.137sow.www',
      '000028': 'ang.154wg.www',
      '000029': 'ang.109aw.www',
      '00002A': 'ang.136aw.www',
      '00002B': 'ang.134arw.www',
      '00002C': 'ang.148fw.www',
      '00002D': 'ang.116acw.www',
      '00002E': 'ang.132dwing.www',
      '00002F': 'ang.119wg.www',
      '000030': 'ang.161arw.www',
      '000031': 'honorguard.www',
      '000032': 'acc.15af.www',
      '000033': 'ang.128arw.www',
      '000034': 'ang.111attackwing.www',
      '000035': '53rdwing.www',
      '000036': 'acc.505ccw.www',
      '000037': 'ang.178wing.www',
      '000038': 'ang.117arw.www',
      '000039': 'ang.103aw.www',
      '00003A': 'afnwc.www',
      '00003B': 'ang.177fw.www',
      '00003C': 'afrc.926wing.www',
      '00003D': 'ang.104fw.www',
      '00003E': 'ang.144fw.www',
      '00003F': 'acc.552acw.www',
      '000040': 'ang.167aw.www',
      '000041': 'pope.www',
      '000042': 'ang.102iw.www',
      '000043': 'ang.131bw.www',
      '000044': 'ang.140wg.www',
      '000045': 'ang.166aw.www',
      '000046': 'ang.106rqw.www',
      '000047': 'ang.146aw.www',
      '000048': 'ang.162wing.www',
      '000049': 'woundedwarrior.www',
      '00004A': 'afsfc.www',
      '00004B': 'ang.180fw.www',
      '00004C': 'ang.129rqw.www',
      '00004D': 'ang.193sow.www',
      '00004E': 'osi.www',
      '00004F': 'learningprofessionals.www',
      '000050': 'ang.175wg.www',
      '000051': 'retirees.www',
      '000052': 'ang.176wg.www',
      '000053': 'afrc.4af.www',
      '000054': 'ang.114fw.www',
      '000055': 'ang.192wg.www',
      '000056': 'ang.157arw.www',
      '000057': 'ang.145aw.www',
      '000058': 'ang.133aw.www',
      '000059': 'ang.182aw.www',
      '00005A': 'ang.173fw.www',
      '00005B': 'afrc.514amw.www',
      '00005C': 'afrc.913ag.www',
      '00005D': 'ang.139aw.www',
      '00005E': 'afrc.940arw.www',
      '00005F': 'hq.safia.www',
      '000060': 'ang.169fw.www',
      '000061': 'afrc.908aw.www',
      '000062': 'acc.1af.www',
      '000063': 'ang.142wg.www',
      '000064': 'ang.185arw.www',
      '000065': '8af.www',
      '000066': '20af.www',
      '000067': 'airmanmagazine.www',
      '000068': 'afhra.www',
      '000069': 'afrc.927arw.www',
      '00006A': 'afrc.944fw.www',
      '00006B': 'afrc.301fw.www',
      '00006C': 'ang.188wg.www',
      '00006D': '37trw.www',
      '00006E': 'creech.www',
      '00006F': 'afrc.pittsburgh.www',
      '000070': 'afrc.419fw.www',
      '000071': 'ang.123aw.www',
      '000072': 'amc.18af.www',
      '000073': '16af.www',
      '000074': 'afrc.459arw.www',
      '000075': 'afrc.442fw.www',
      '000076': 'afrc.westover.www',
      '000077': 'ang.angtec.www',
      '000078': 'afrc.niagara.www',
      '000079': 'afrc.507arw.www',
      '00007A': 'afsc.www',
      '00007B': 'safety.www',
      '00007C': 'afrc.302aw.www',
      '00007D': 'afrc.919sow.www',
      '00007E': 'afrc.10af.www',
      '00007F': 'aetc.torch.www',
      '000080': 'afrc.624rsg.www',
      '000081': 'afrc.920rqw.www',
      '000082': 'aftc.www',
      '000083': 'afrc.931arw.www',
      '000084': 'afrc.youngstown.www',
      '000085': 'afrc.march.www',
      '000086': 'mortuary.www',
      '000087': 'afrc.minneapolis.www',
      '000088': 'afcec.www',
      '000089': 'aflcmc.www',
      '00008A': 'beale.www',
      '00008B': 'afsoc.www',
      '00008C': 'afrc.arpc.www',
      '00008D': 'vance.www',
      '00008E': 'ang.127wg.www',
      '00008F': 'afrc.403wg.www',
      '000090': 'afrc.446aw.www',
      '000091': 'afdw.www',
      '000092': 'pacaf.7af.www',
      '000093': 'afrc.932aw.www',
      '000094': 'usafe.501csw.www',
      '000095': 'afrc.349amw.www',
      '000096': 'maxwell.www',
      '000097': 'arnold.www',
      '000098': 'bmtflightphotos.www',
      '000099': 'afrc.433aw.www',
      '00009A': 'afrc.dobbins.www',
      '00009B': 'hurlburt.www',
      '00009C': 'cannon.www',
      '00009D': 'afrc.grissom.www',
      '00009E': 'afrc.307bw.www',
      '00009F': 'tinker.www',
      '0000A0': 'mcchord.www',
      '0000A1': 'afrc.445aw.www',
      '0000A2': 'laughlin.www',
      '0000A3': 'afrc.512aw.www',
      '0000A4': 'altus.www',
      '0000A5': 'robins.www',
      '0000A6': 'pacaf.5af.www',
      '0000A7': 'kirtland.www',
      '0000A8': '33fw.www',
      '0000A9': 'sheppard.www',
      '0000AA': 'hill.www',
      '0000AB': 'mcconnell.www',
      '0000AC': 'luke.www',
      '0000AD': 'malmstrom.www',
      '0000AE': 'jba.www',
      '0000AF': 'mountainhome.www',
      '0000B0': 'ellsworth.www',
      '0000B1': 'macdill.www',
      '0000B2': 'incirlik.www',
      '0000B3': 'hanscom.www',
      '0000B4': 'seymourjohnson.www',
      '0000B5': 'tyndall.www',
      '0000B6': 'eielson.www',
      '0000B7': 'littlerock.www',
      '0000B8': 'lakenheath.www',
      '0000B9': 'whiteman.www',
      '0000BA': 'holloman.www',
      '0000BB': 'fairchild.www',
      '0000BC': 'misawa.www',
      '0000BD': 'afgsc.www',
      '0000BE': 'minot.www',
      '0000BF': 'offutt.www',
      '0000C0': 'afrc.315aw.www',
      '0000C1': 'grandforks.www',
      '0000C2': 'scott.www',
      '0000C3': 'andersen.www',
      '0000C4': 'shaw.www',
      '0000C5': 'goodfellow.www',
      '0000C6': 'warren.www',
      '0000C7': 'barksdale.www',
      '0000C8': 'spangdahlem.www',
      '0000C9': 'dyess.www',
      '0000CA': 'dover.www',
      '0000CB': 'eglin.www',
      '0000CC': 'afrc.www',
      '0000CD': 'amc.www',
      '0000CE': 'aviano.www',
      '0000CF': 'kadena.www',
      '0000D0': 'music.www',
      '0000D1': 'afmc.www',
      '0000D2': 'moody.www',
      '0000D3': 'edwards.www',
      '0000D4': 'ramstein.www',
      '0000D5': 'pacaf.www',
      '0000D6': 'osan.www',
      '0000D7': 'kunsan.www',
      '0000D8': 'mildenhall.www',
      '0000D9': 'acc.www',
      '0000DA': 'keesler.www',
    },
    name: 'af',
  },
  '000002': {
    children: {
      '000001': 'www',
      '000002': 'cascom',
      '000003': 'usarj.www',
      '000004': 'cyberdefensereview',
      '000005': '1tsc.www',
      '000006': 'arnorth.www',
      '000007': 'cyber',
      '000008': 'jpeoaa',
      '000009': 'amcom.www',
      '00000A': 'usafmcom.www',
      '00000B': 'tobyhanna.www',
      '00000C': 'afsbeurope.www',
      '00000D': 'arcp.www',
      '00000E': 'psmagazine.www',
      '00000F': 'ngb.il.www',
      '000010': 'first.www',
      '000011': 'aschq.www',
      '000012': 'armyupress.www',
      '000013': 'europeafrica.www',
    },
    name: 'army',
  },
  '000003': {
    children: {
      '000001': 'www',
    },
    name: 'ctoinnovation',
  },
  '000004': {
    children: {
      '000001': 'www',
    },
    name: 'cybercom',
  },
  '000005': {
    children: {
      '000001': 'www',
    },
    name: 'dantes',
  },
  '000006': {
    children: {
      '000001': 'www',
    },
    name: 'dcaa',
  },
  '000007': {
    children: {
      '000001': 'www',
    },
    name: 'dcma',
  },
  '000008': {
    children: {
      '000001': 'www',
    },
    name: 'dfas',
  },
  '000009': {
    children: {
      '000001': 'www',
    },
    name: 'disa',
  },
  '00000A': {
    children: {
      '000001': 'www',
    },
    name: 'dla',
  },
  '00000B': {
    children: {
      '000001': 'defensetravel.www',
      '000002': 'travel.www',
    },
    name: 'dod',
  },
  '00000C': {
    children: {
      '000001': 'www',
    },
    name: 'dodig',
  },
  '00000D': {
    children: {
      '000001': 'plainsguardian',
      '000002': 'navylive.jag',
      '000003': 'minationalguard',
      '000004': 'navylive.usnhistory',
      '000005': 'navylive.seabeemagazine',
    },
    name: 'dodlive',
  },
  '00000E': {
    children: {
      '000001': 'ndia',
    },
    name: 'dtic',
  },
  '00000F': {
    children: {
      '000001': 'eitpmo',
      '000002': 'usamriid',
      '000003': 'phcp',
      '000004': 'usariem',
      '000005': 'usammda',
      '000006': 'wrair',
      '000007': 'blastinjuryresearch',
      '000008': 'jts',
      '000009': 'medicalmuseum',
      '00000A': 'mrdc',
    },
    name: 'health',
  },
  '000010': {
    children: {
      '000001': 'navydsrc',
      '000002': 'erdc.www',
      '000003': 'afrl',
      '000004': 'arl',
      '000005': 'centers',
    },
    name: 'hpc',
  },
  '000011': {
    children: {
      '000001': 'jbab.www',
      '000002': 'jber.www',
    },
    name: 'jb',
  },
  '000012': {
    children: {
      '000001': 'www',
    },
    name: 'jcs',
  },
  '000013': {
    children: {
      '000001': 'www',
    },
    name: 'jtnc',
  },
  '000014': {
    children: {
      '000001': 'www',
      '000002': 'marsoc.www',
      '000003': 'mcieast.www',
      '000004': 'mcbbutler.www',
      '000005': 'mciwest.www',
      '000006': '6thmcd.www',
      '000007': 'japan.www',
      '000008': '3rdmardiv.www',
      '000009': '9thmcd.www',
      '00000A': 'mcasiwakunijp.www',
      '00000B': '3rdmlg.www',
      '00000C': 'beaufort.www',
      '00000D': 'albany.www',
      '00000E': '15thmeu.www',
      '00000F': 'marineband.www',
      '000010': 'miramar-ems.www',
      '000011': 'trngcmd.www',
      '000012': '2ndmlg.www',
      '000013': 'marforpac.www',
      '000014': '3rdmaw.www',
      '000015': 'mcasiwakuni.www',
    },
    name: 'marines',
  },
  '000015': {
    children: {
      '000001': 'www',
    },
    name: 'metc',
  },
  '000016': {
    children: {
      '000001': 'arkansas',
    },
    name: 'nationalguard',
  },
  '000017': {
    children: {
      '000001': 'jag.www',
      '000002': 'cnrc.www',
      '000003': 'navair.frcsw',
      '000004': 'usff.cnmoc.www',
      '000005': 'navyreserve.www',
      '000006': 'usff.sublant.www',
      '000007': 'amphib7flt.www',
      '000008': 'nrl.www',
      '000009': 'airpac.www',
      '00000A': 'navfac.www',
      '00000B': 'usff.msc.www',
      '00000C': 'cnic.cnrnw',
      '00000D': 'nepa.www',
      '00000E': 'usff.surflant.www',
      '00000F': 'usff.airlant.www',
      '000010': 'c6f.www',
      '000011': 'netc.www',
      '000012': 'surfpac.www',
      '000013': 'usff.www',
      '000014': 'cpf.www',
      '000015': 'www',
      '000016': 'allhands',
      '000017': 'navsea.www',
      '000018': 'mynavyhr.www',
    },
    name: 'navy',
  },
  '000018': {
    children: {
      '000001': 'me.www',
      '000002': 'scguard.www',
      '000003': 'ok',
      '000004': 'nh',
      '000005': 'dc',
      '000006': 'co',
      '000007': 'wv.www',
      '000008': 'pa.www',
      '000009': 'public.vt',
      '00000A': 'ak',
      '00000B': 'ky',
      '00000C': 'va',
    },
    name: 'ng',
  },
  '000019': {
    children: {
      '000001': 'moguard.www',
    },
    name: 'ngb',
  },
  '00001A': {
    children: {
      '000001': 'dote.www',
    },
    name: 'osd',
  },
  '00001B': {
    children: {
      '000001': 'www',
    },
    name: 'pacom',
  },
  '00001C': {
    children: {},
    name: 'serdp-estcp',
  },
  '00001D': {
    children: {
      '000001': 'jtfb.www',
      '000002': 'www',
    },
    name: 'southcom',
  },
  '00001E': {
    children: {
      '000001': 'spoc.www',
      '000002': 'www',
      '000003': 'losangeles.www',
      '000004': 'buckley.www',
    },
    name: 'spaceforce',
  },
  '00001F': {
    children: {
      '000001': 'landstuhl',
      '000002': 'belvoirhospital',
      '000003': 'portsmouth',
      '000004': 'tripler',
      '000005': 'newsroom',
    },
    name: 'tricare',
  },
  '000020': {
    children: {
      '000001': 'pacificarea.www',
      '000002': 'atlanticarea.www',
      '000003': 'dco.www',
      '000004': 'www',
      '000005': 'news.www',
      '000006': 'dcms.www',
      '000007': 'mycg.www',
    },
    name: 'uscg',
  },
  '000021': {
    children: {
      '000001': 'www',
      '000002': 'esd.www',
    },
    name: 'whs',
  },
};

assert assertion.validateDomains(domains);

{
  domains: domains,
}
