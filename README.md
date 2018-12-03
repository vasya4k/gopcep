
 OPEN Object-Class is 1.

    OPEN Object-Type is 1.

    The format of the OPEN object body is as follows:

     0                   1                   2                   3
     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    | Ver |   Flags |   Keepalive   |  DeadTimer    |      SID      |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
    |                                                               |
    //                       Optional TLVs                         //
    |                                                               |
    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

                     Figure 9: OPEN Object Format
 As of today 02.12.2018 and testing against Juniper vMX JunOS 17.2R1.13
 If you see Common Object Header length is 24 Bytes 4 bytes is the CommonObjectHeader
 next 4 bytes is OPEN Object so it is 24-4-4 = 16. The remainig 16 are Optional TLVs and can be found
 In PCEP extensions described in https:tools.ietf.org/html/rfc8231#section-7.1.1
 Path Computation Element Communication Protocol (PCEP) Extensions for Stateful PCE

 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |               Type=16         |            Length=4           |
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 |                             Flags                           |U|
 +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

 		 Figure 9: STATEFUL-PCE-CAPABILITY TLV Format

 The type (16 bits) of the TLV is 16.  The length field is 16 bits
 long and has a fixed value of 4
 The value comprises a single field -- Flags (32 bits):

    U (LSP-UPDATE-CAPABILITY - 1 bit):  if set to 1 by a PCC, the U flag
       indicates that the PCC allows modification of LSP parameters; if
       set to 1 by a PCE, the U flag indicates that the PCE is capable of

 That gives us another 2+2+4 = 8 bytes so 16-8 = 8 bytes remaining and
 we need to look into another rfc draft https:tools.ietf.org/html/draft-ietf-pce-segment-routing-14#section-5.1
 0                   1                   2                   3
        0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |            Type=26            |            Length=4           |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
       |         Reserved              |   Flags   |N|L|      MSD      |
       +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

                 Figure 1: SR-PCE-CAPABILITY sub-TLV format
 The code point for the TLV type is 26.  The TLV length is 4 octets.
 The type (16 bits) The length field is 16 bits.
 The 32-bit value is formatted as follows.

    Reserved:  MUST be set to zero by the sender and MUST be ignored by
       the receiver.

    Flags:  This document defines the following flag bits.  The other
       bits MUST be set to zero by the sender and MUST be ignored by the
       receiver.

       *  N: A PCC sets this bit to 1 to indicate that it is capable of
          resolving a Node or Adjacency Identifier (NAI) to a SID.

       *  L: A PCC sets this bit to 1 to indicate that it does not impose
          any limit on the MSD.

    Maximum SID Depth (MSD):  specifies the maximum number of SIDs (MPLS
       label stack depth in the context of this document) that a PCC is
       capable of imposing on a packet.  Section 6.1 explains the
       relationship between this field and the L bit.
 The above object is  2+2+4 = 8 bytes and now we can see this adds up
 if we test agains a juniper device.

 Whole MSG: [00100000 00000001 00000000 00011100 00000001 00010000 00000000 00011000 00100000 00011110 01111000 00100010 00000000 00010000 00000000 00000100 00000000 00000000 00000000 00000101 00000000 00011010 00000000 00000100 00000000 00000000 00000000 00000101] 
{
  "Version": 1,
  "Flags": 0,
  "MessageType": 1,
  "MessageLength": 28
}
{
  "ObjectClass": 1,
  "ObjectType": 1,
  "Reservedfield": 0,
  "ProcessingRule": false,
  "Ignore": false,
  "ObjectLength": 24
}
{
  "Version": 1,
  "Flags": 0,
  "Keepalive": 30,
  "DeadTimer": 120,
  "SID": 34
}
UFlag: 00000001 
{
  "Type": 16,
  "Length": 4,
  "Flags": 5,
  "UFlag": true
}
SR Cap: [00000000 00011010 00000000 00000100 00000000 00000000 00000000 00000101] 
{
  "Type": 26,
  "Length": 4,
  "Reserved": 0,
  "NAIToSID": false,
  "NoMSDLimit": false,
  "MSD": 5
}
