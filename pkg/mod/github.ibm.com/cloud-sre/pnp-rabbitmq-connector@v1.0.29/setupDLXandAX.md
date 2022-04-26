# How to set up DLX and AX in PnP RabbitMQ
---

DLX = Dead Letter Exchange  
AX = Alternate Exchange  

1. Access the management UI URL with your browser, and login  

2. Click the `Exchanges` tab, if not there already  

3. Add both exchanges, at the moment they are called `pnp.direct.deadletter` for DLX and `pnp.direct.alternate` for AX.  

    - **For DLX**  
    Name: pnp.direct.deadletter  
    Type: fanout  
    Durability: Durable  
    Auto delete: No  
    Internal: Yes  

    Click `Add Exchange`  

    - **For AX**  
    Name: pnp.direct.alternate  
    Type: fanout  
    Durability: Durable  
    Auto delete: No  
    Internal: Yes  

    Click `Add Exchange`  
    Both exchanges should have been added, confirm they have the Features D and I.  

4. Click the `Queues` tab  

5. Add queues `pnp.deadletter.msgs` and `pnp.unrouted.msgs`  

    - **For pnp.deadletter.msgs**  
    Name: pnp.deadletter.msgs  
    Durability: Durable  
    Auto delete: No  

    Click `Add Queue`  

    - **For pnp.unrouted.msgs**  
    Name: pnp.unrouted.msgs  
    Durability: Durable  
    Auto delete: No  

    Click `Add Queue`  
    Both queues should have been added

6. Configure queues just added

    - Click the `pnp.deadletter.msgs` queue, in the list.  
    - Under Bindings, add `pnp.direct.deadletter` in the `From exchange` field.  
    - Click Bind.  
    - For `pnp.unrouted.msgs` queue, do the same but use `pnp.direct.alternate` in the `From exchange` field.  

    Now we have the queues bound to the exchanges created.  

7. Click the `Admin` tab, then `Policies`  

8. Add policies for `DLX` and `AX`, under `Add / update a policy`  

    - **For DLX**  
    Name: DLX  
    Pattern: .*  
    Apply to: Queues  
    Priority: 1  
    Definition: dead-letter-exchange = pnp.direct.deadletter  

    - **For AX**  
    Name: AX  
    Pattern: ^pnp.direct$  
    Apply to: Exchanges  
    Priority: 2  
    Definition: alternate-exchange = pnp.direct.alternate  

    Note: for the `Priority` field, all the policies in the list must have different priorities for them to work.  

9. Confirm policies have been applied  

    - In the `Queues` tab, all queues should have `DLX` in the Features.  
    - In the `Exchanges` tab, the exchange `pnp.direct` should have `AX` in the Features.  

---

# How to set up DLX and AX for EDB in PnP RabbitMQ
---

EDBDLX = Dead Letter Exchange  
EDBAX = Alternate Exchange  

1. Access the management UI URL with your browser, and login  

2. Click the `Exchanges` tab, if not there already  

3. Add both exchanges, at the moment they are called `edb.direct.deadletter` for EDBDLX and `edb.direct.alternate` for EDBAX.  

    - **For EDBDLX**  
    Name: edb.direct.deadletter  
    Type: fanout  
    Durability: Durable  
    Auto delete: No  
    Internal: Yes  

    Click `Add Exchange`  

    - **For EDBAX**  
    Name: edb.direct.alternate  
    Type: fanout  
    Durability: Durable  
    Auto delete: No  
    Internal: Yes  

    Click `Add Exchange`  
    Both exchanges should have been added, confirm they have the Features D and I.  

4. Click the `Queues` tab  

5. Add queues `edb.deadletter.msgs` and `edb.unrouted.msgs`  

    - **For edb.deadletter.msgs**  
    Name: edb.deadletter.msgs  
    Durability: Durable  
    Auto delete: No  

    Click `Add Queue`  

    - **For edb.unrouted.msgs**  
    Name: edb.unrouted.msgs  
    Durability: Durable  
    Auto delete: No  

    Click `Add Queue`  
    Both queues should have been added

6. Configure queues just added

    - Click the `edb.deadletter.msgs` queue, in the list.  
    - Under Bindings, add `edb.direct.deadletter` in the `From exchange` field.  
    - Click Bind.  
    - For `edb.unrouted.msgs` queue, do the same but use `edb.direct.alternate` in the `From exchange` field.  

    Now we have the queues bound to the exchanges created.  

7. Click the `Admin` tab, then `Policies`  

8. Add policies for `EDBDLX` and `EDBAX`, under `Add / update a policy`  

    - **For EDBDLX**  
    Name: EDBDLX  
    Pattern: ^edb.*  
    Apply to: Queues  
    Priority: 3  
    Definition: dead-letter-exchange = edb.direct.deadletter  

    - **For EDBAX**  
    Name: EDBAX  
    Pattern: ^edb.direct$  
    Apply to: Exchanges  
    Priority: 4  
    Definition: alternate-exchange = edb.direct.alternate  

    Note: for the `Priority` field, all the policies in the list must have different priorities for them to work.  

9. Confirm policies have been applied  

    - In the `Queues` tab, all EDB queues should have `EDBDLX` in the Features.  
    - In the `Exchanges` tab, the exchange `edb.direct` should have `EDBAX` in the Features.  
    - The final policy page should look like this: ![PolicySetup](_examples/RabbitMQPolicySetup.png)
