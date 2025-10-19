6.2.3.2 EAS Discovery Procedure
6.2.3.2.1 General
For PDU Session with Session Breakout connectivity model, based on UE subscription (e.g. DNN) and/or the operator's
configuration, the DNS Query sent by UE may be handled by an EASDF (see clause 6.2.3.2.2), or by a local or central
DNS resolver/server (see clause 6.2.3.2.3).
NOTE: For the scenario where the TE and MT are separated, information provided by the SMF in the NAS
message during the PDU Session Establishment or Modification may not be provided to the TE. Annex C
documents mitigations for this scenario.
6.2.3.2.2 EAS Discovery Procedure with EASDF
For the case that the UE DNS Query is to be handled by EASDF, the following applies.
- The AF may provide EAS Deployment Information to NEF which may store it in UDR, as defined in
clause 6.2.3.4. SMF may retrieve EAS Deployment Information from NEF as described in clause 6.2.3.4 or has
locally preconfigured information. EAS Deployment Information is used for creating DNS message handling
rule on EASDF and it is not dedicated to specific UE session(s).
 EAS Deployment Information may apply to all PDU Sessions with a certain DNN, S-NSSAI and/or specific
Internal Group Identifier(s).
- The SMF may provide BaselineDNSPattern to EASDF, the BaselineDNSPattern are derived from EAS
Deployment Information provided by AF and are not dedicated to specific PDU Session; SMF configures
EASDF with BaselineDNSPattern according to the procedures defined in clause 6.2.3.4.
 The Baseline DNS message detection template ID may be used by the EASDF to refer to Baseline DNS message
detection template, and derive array of FQDN ranges and/or array of EAS IP address ranges. The Baseline DNS
handling actions ID may be used by the EASDF to refer to Baseline DNS handling actions information, and
derive actions related parameters.
 The Baseline DNS message detection template ID and the Baseline DNS handling actions ID are unique per
SMF set when a SMF set controls an EASDF and shall be unique per SMF otherwise, within an EASDF
Baseline
 BaselineDNSPattern may contain one or several items, where each item is either a Baseline DNS message
detection template or a Baseline DNS handling actions information. Each BaselineDNSPattern item may be
updated or deleted using Baseline DNS message detection template ID or Baseline DNS handling actions ID to
identify the updated or deleted item
- Baseline DNS message detection template
- Baseline DNS message detection template ID
- DNS message type = DNS Query or DNS Response:
- If DNS message type = DNS Query:
- Array of (FQDN ranges).
- If DNS message type = DNS Response:
- Array of FQDN ranges and/or array of EAS IP address ranges.
- Baseline DNS handling actions information:
- Baseline DNS handling actions ID:
- ECS option.
- Local DNS server IP address.
NOTE 1: The FQDN can be set to wildcard to indicate the default DNS Server (e.g. the C-DNS), for the case in
which the DNS message should be forwarded to the default DNS Server.
NOTE 2: The BaselineDNSPattern can be configured for a specific application with the related FQDN set in the
detection template.
NOTE 3: The definition of structure of Baseline DNS handling actions ID and Detection template ID is left to
stage 3. As an example, Baseline DNS handling action ID and Detection template ID could contain a
concatenation of the SMF ID or SMF set Id and of SMF implementation selected information such as the
DNAI or a sequence number. The EASDF is not meant to understand the structure of Baseline DNS
handling actions ID and Detection template ID.
- During the PDU Session establishment procedure, the SMF may obtain the EAS Deployment Information from
the NEF if not already retrieved (by subscription of such information to the NEF as described in clause 6.2.3.4.3)
or the SMF is preconfigure with the EAS Deployment Information and the SMF selects an EASDF and provides
its address to the UE as the DNS Server to be used for the PDU Session.
 The SMF configures the EASDF with DNS message handling rules to handle DNS messages related to the
UE(s). The DNS message handling rule has a unique identifier and includes information used for DNS message
detection and associated action(s). The DNS handling rules is defined as following:
- Precedence of the DNS message handling rule;
- DNS Handling Rule Identity;
- A Baseline DNS message detection template ID and/or a DNS message detection template (optional and
includes at least one of the following, if existing):
- DNS message type = DNS Query or DNS Response:
- If DNS message type = DNS Query:
- Source IP address (i.e. UE IP address).
- Array of (FQDN ranges) (optional).
- If DNS message type = DNS Response:
- Array of FQDN ranges and/or array of EAS IP address ranges (optional).
- DNS message Identifier (if received from EASDF);
NOTE 4: For DNS message type = Query, the UE IP address provided at DNS context creation
(Neasdf_DNSContext_Create Request) is considered if not provided explicitly as part of the DNS
message detection template.
NOTE 5: DNS message Identifier is used by EASDF for matching between the message reported in the
Neasdf_DNSContext_Notify and the corresponding DNS message handling rule included in
Neasdf_DNSContext_Update.
- Action(s) (includes at least one action); the possible actions include:
- Reporting Action: Report DNS message content to SMF (i.e. target FQDN and if available: IP address
information provided back by the DNS server). This reporting action may include reporting-once
indication. If this indication is included, the EASDF reports the DNS message content to the SMF once if
the DNS message detection template matches the first incoming DNS Query or DNS Response message. 
NOTE 6: With reporting-once indication, the DNS message detection template should contain the EAS IP address
ranges corresponding to the same DNAI. Resetting the Reporting-once indication can be used by the SMF
to allow reporting associated with a DNS handling rule when the SMF has removed the UL-CL/BP e.g.
when the UE has moved out of the area associated with the current DNAI and thus insertion of a new
UPF offloading capability can be considered.
- Forwarding Action: Send the DNS message(s) to a DNS server/resolver(s) as follows:
A. (possibly) Including the information to build optional EDNS Client Subnet option in the DNS
message (The information for the EASDF to build the EDNS Client Subnet option is either included
in the DNS handling rule, or Baseline DNS handling actions ID acts as a reference to the Baseline
DNS handling actions Information. This corresponds to the option A defined below.
B. the information for the DNS message target address is either included as DNS Server Address
indicated in the DNS handling rule, or the Baseline DNS handling actions ID included in the DNS
handling rules refers to DNS message target address information; if no DNS Server Address is
provided by the SMF in the rule, then the EASDF is to forward the DNS message to a locally
preconfigured default DNS server/resolver. This corresponds to the option B defined below.
NOTE 7: The forwarding action can include either A or B.
- Control Action: Performs at least one of control actions on the DNS message(s) as follows:
- Buffer the DNS message(s).
- Send the buffered DNS Response(s) message to UE.
- Discard cached DNS Response message(s).
When the EASDF forwards a DNS message (to the UE or towards a DNS server over N6), it uses its own address as the
source address of the DNS message.
The SMF may use following information to create DNS message handling rules associated with a PDU Session:
- Local configuration associated with the (DNN, S-NSSAI, Internal Group Identifier) of the PDU Session; and/or
- EAS Deployment Information provided by the AF or preconfigured in the SMF; and/or
- Information derived from the UE location such as candidate L-PSA(s); and/or
- PDU Session information, like PDU Session L-PSA(s) and ULCL/BP; and/or
- Internal Group Identifier received in the Session Management Subscription data from the UDM;
NOTE 7: For example, the SMF can derive the IP address for ECS based on the N6 IP address(es) associated with
serving L-PSA(s) locally configured or in the NRF.
NOTE 8: Providing in DNS EDNS Client Subnet option an IP address associated with the L-PSA UPF protects the
privacy of the (IP address of the) UE.
- If the FQDN in a DNS Query matches the FQDN(s) provided by the SMF in a DNS message detection template,
based on instructions by SMF, one of the following options is executed by the EASDF based on a corresponding
DNS message handling rule:
- Option A: The EASDF includes an EDNS Client Subnet (ECS) option into the DNS Query message as
defined in RFC 7871[6] and sends the DNS Query message to the DNS server for resolving the FQDN. The
DNS server may resolve the EAS IP address considering the EDNS Client Subnet option and sends the DNS
Response to the EASDF;
- Option B: The EASDF sends the DNS Query message to a Local DNS server which is responsible for
resolving the FQDN within the corresponding L-DN. The EASDF receives the DNS Response message from
the Local DNS server.
NOTE 9: Option B does not support the scenario where the PSA UPF for transferring DNS Query between EASDF
and DNS server, or the EASDF has no direct connectivity with the Local DNS servers. 
The SMF instructions for a matching FQDN may as well indicate EASDF to contact SMF. SMF then provides
the EASDF with a DNS message handling rule;
- If the DNS Query from the UE does not match a DNS message handling rules set by the SMF, then the EASDF
may simply forward the DNS Query towards a preconfigured DNS server/resolver for DNS resolution;
- When the EASDF receives a DNS Response message, the EASDF notifies the EAS information (i.e. EAS IP
address(es), the EAS FQDN and if available the corresponding IP address within the ECS DNS option) to the
SMF if the DNS message reporting condition provided by the SMF is met (i.e. the EAS IP address or FQDN is
within the IP/FQDN range). The SMF may then select the target DNAI based on the EAS information and
trigger UL CL/BP and L-PSA insertion as specified in clause 6.3.3 in TS 23.501 [2] based on the Notification.
NOTE 10: To avoid SMF overloading caused by massive reporting, the overload control mechanisms defined in
clause 6.4 of TS 29.500 [9] can be used.
 The information to build the EDNS Client Subnet option or the Local DNS server address provided by the SMF
to the EASDF are part of the DNS message handling rules to handle DNS Queries from the UE. This
information is related to DNAI(s) for that FQDN(s) for the UE location. The SMF may provide DNS message
handling rules to handle DNS Queries from the UE to the EASDF when the SMF establishes the association with
the EASDF for the UE and may update the rules at any time when the association exists. For the selection of the
candidate DNAI for a FQDN for the UE, the SMF may consider the UE location, network topology, EAS
Deployment Information and related policy information for the PDU Session provided as defined in
TS 23.503 [4] clause 6.4 or be preconfigured into the SMF. After the UE mobility, if the provided Information
for EDNS Client Subnet option or the Local DNS server address needs to be updated, the SMF may send an
update of DNS message handling rules to the EASDF.
NOTE 11: If multiple candidate DNAIs are available after considering the UE location, network topology and EAS
deployment, the SMF selects one DNAI from the multiple ones based on operator's policy. For examples,
the SMF can select the DNAI randomly, or based on selection weight factor if provided by AF, or select
the DNAI closest to the UE location.
NOTE 12: To protect the SMF (e.g. to block DOS from the EASDF), the EASDF IP address for DNS Query Request
is only accessible from the UE IP address via UPF.
Once the UL CL/BP and L-PSA have been inserted, the SMF may decide that the DNS messages for the FQDN are to
be handled by Local DNS resolver/server from now on. This option is further described in clause 6.2.3.2.3.
To avoid EASDF sending redundant DNS message reports triggering UL CL/BP insertion corresponding to the same
DNAI, the SMF may send reporting-once control information (i.e. DNS message handling rule with DNS message
detection template containing EAS IP address ranges with reporting-once indication set) to EASDF to instruct the
EASDF to report only once for the DNS messages matching with the DNS message detection template of the reportingonce control information for the DNS message detection template. In addition, the SMF may instruct the EASDF not to
report DNS Responses to SMF corresponding to some FQDN ranges and/or EAS IP address ranges e.g. once the UL
CL/BP and L-PSA have been inserted for the corresponding EAS IP address ranges for Pre-established session breakout
while there is configuration for the related EASDF reporting DNS Responses. After the removal or change of the LPSA, the SMF may instruct the EASDF to restart the reports of the DNS messages.
If the SMF, based on local configuration, decides that the interaction between EASDF and DNS Server in the DN shall
go via an UPF, the SMF sends corresponding N4 rules to this UPF to instruct this UPF to forward DNS message
between EASDF and the external DNS server. In this case, DNS messages between EASDF and DNS Server described
in this clause are transferred via this UPF transparently.
NOTE 13: Based network configuration, one UPF is used to transmit DNS signalling between EASDF and DNS
servers. The N4 session between the SMF and this UPF is not related to a specific PDU Session but
provides rules targeting Downlink traffic from DNS servers to the EASDF and associated with the traffic
of multiple UE(s); the traffic forwarding between EASDF and this UPF is realized by IP in IP tunnelling
.The EASDF provides the SMF with the source address it uses to contact DNS servers and with the
destination address where it expects to receive the tunnelled traffic. 