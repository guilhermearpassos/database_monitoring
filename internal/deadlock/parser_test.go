package deadlock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const reverseOrderXML = `<deadlock>
  <victim-list>
    <victimProcess id="p1" />
  </victim-list>
  <process-list>
    <process id="p1" logused="2416" waitresource="KEY: 6:72057594058113024 (8194)" waittime="3412" lockMode="X" status="suspended" spid="64" ecid="0" priority="0" trancount="2" transactionname="user_transaction" clientapp="OrdersAPI" hostname="aks-orders" loginname="svc_orders" isolationlevel="read committed (2)" currentdb="6" currentdbname="OrdersDB">
      <executionStack>
        <frame procname="OrdersDB.dbo.usp_ReserveStock" line="42" sqlhandle="0x01">UPDATE dbo.Inventory SET Qty = Qty - 1 WHERE SkuId = @id;</frame>
      </executionStack>
      <inputbuf>EXEC dbo.usp_SubmitOrder @OrderId = 10;</inputbuf>
    </process>
    <process id="p2" logused="18640" waitresource="KEY: 6:72057594041335808 (c401)" waittime="2980" lockMode="U" status="suspended" spid="71" ecid="0" priority="0" trancount="2" transactionname="user_transaction" clientapp="InventoryWorker" hostname="aks-worker" loginname="svc_inv" isolationlevel="read committed (2)" currentdb="6" currentdbname="OrdersDB">
      <executionStack>
        <frame procname="OrdersDB.dbo.usp_Reconcile" line="88" sqlhandle="0x02">UPDATE dbo.Orders SET Status = 'ALLOC' WHERE OrderId = @id;</frame>
      </executionStack>
      <inputbuf>EXEC dbo.usp_Reconcile;</inputbuf>
    </process>
  </process-list>
  <resource-list>
    <keylock hobtid="72057594058113024" dbid="6" objectname="OrdersDB.dbo.Inventory" indexname="PK_Inventory" id="l1" mode="X">
      <owner-list><owner id="p2" mode="X" /></owner-list>
      <waiter-list><waiter id="p1" mode="X" requestType="wait" /></waiter-list>
    </keylock>
    <keylock hobtid="72057594041335808" dbid="6" objectname="OrdersDB.dbo.Orders" indexname="PK_Orders" id="l2" mode="X">
      <owner-list><owner id="p1" mode="X" /></owner-list>
      <waiter-list><waiter id="p2" mode="U" requestType="wait" /></waiter-list>
    </keylock>
  </resource-list>
</deadlock>`

const bookmarkLookupXML = `<deadlock-list><deadlock>
  <victim-list><victimProcess id="p1" /></victim-list>
  <process-list>
    <process id="p1" spid="88" waittime="1850" logused="0" lockMode="S" clientapp="ReportingSvc" isolationlevel="read committed (2)" currentdbname="OrdersDB">
      <executionStack><frame procname="adhoc" line="1">SELECT * FROM dbo.Orders WHERE Status = @p0;</frame></executionStack>
    </process>
    <process id="p2" spid="91" waittime="1620" logused="4120" lockMode="X" clientapp="CheckoutAPI" isolationlevel="read committed (2)" currentdbname="OrdersDB">
      <executionStack><frame procname="OrdersDB.dbo.usp_Pay" line="14">UPDATE dbo.Orders SET Status = 'PAID' WHERE OrderId = @id;</frame></executionStack>
    </process>
  </process-list>
  <resource-list>
    <keylock hobtid="101" dbid="6" objectname="OrdersDB.dbo.Orders" indexname="PK_Orders" id="l1" mode="X">
      <owner-list><owner id="p2" mode="X" /></owner-list>
      <waiter-list><waiter id="p1" mode="S" requestType="wait" /></waiter-list>
    </keylock>
    <keylock hobtid="102" dbid="6" objectname="OrdersDB.dbo.Orders" indexname="IX_Orders_Status" id="l2" mode="S">
      <owner-list><owner id="p1" mode="S" /></owner-list>
      <waiter-list><waiter id="p2" mode="X" requestType="wait" /></waiter-list>
    </keylock>
  </resource-list>
</deadlock></deadlock-list>`

const lockUpgradeXML = `<deadlock>
  <victim-list><victimProcess id="p1" /></victim-list>
  <process-list>
    <process id="p1" spid="55" waittime="4120" logused="640" lockMode="X" clientapp="Billing" isolationlevel="repeatable read (3)" currentdbname="BillingDB">
      <inputbuf>UPDATE dbo.AccountBalances SET Balance = Balance + 10 WHERE AccountId = 1;</inputbuf>
    </process>
    <process id="p2" spid="61" waittime="3990" logused="896" lockMode="X" clientapp="Billing" isolationlevel="repeatable read (3)" currentdbname="BillingDB">
      <inputbuf>UPDATE dbo.AccountBalances SET Balance = Balance - 5 WHERE AccountId = 1;</inputbuf>
    </process>
  </process-list>
  <resource-list>
    <keylock hobtid="201" dbid="7" objectname="BillingDB.dbo.AccountBalances" indexname="PK_AccountBalances" id="l1" mode="S">
      <owner-list>
        <owner id="p1" mode="S" />
        <owner id="p2" mode="S" />
      </owner-list>
      <waiter-list>
        <waiter id="p1" mode="X" requestType="convert" />
        <waiter id="p2" mode="X" requestType="convert" />
      </waiter-list>
    </keylock>
  </resource-list>
</deadlock>`

const exchangeXML = `<deadlock>
  <victim-list><victimProcess id="p1" /></victim-list>
  <process-list>
    <process id="p1" spid="102" ecid="3" waittime="500" currentdbname="AnalyticsDB"><inputbuf>SELECT COUNT(*) FROM dbo.BigFact;</inputbuf></process>
  </process-list>
  <resource-list>
    <exchangeEvent id="Pipe1" WaitType="e_waitPipeGetRow" nodeId="4">
      <owner-list><owner id="p1" mode="e_waitPipeNewRow" /></owner-list>
      <waiter-list><waiter id="p1" mode="e_waitPipeGetRow" /></waiter-list>
    </exchangeEvent>
  </resource-list>
</deadlock>`

func TestParseXMLPatterns(t *testing.T) {
	t.Run("reverse order key lock", func(t *testing.T) {
		p, err := ParseXML(reverseOrderXML)
		require.NoError(t, err)
		require.Equal(t, PatternReverseOrder, p.Diagnosis.Pattern)
		require.Equal(t, "64", p.VictimSPID)
		require.Equal(t, int64(3412), p.MaxWaitMs)
		require.Equal(t, []string{"OrdersDB"}, p.Databases)
		require.Len(t, p.Processes, 2)
		require.True(t, p.Processes[0].IsVictim)
		require.False(t, p.Processes[1].IsVictim)
		require.Len(t, p.Resources, 2)
		require.NotEmpty(t, p.Fingerprint)
		require.NotEmpty(t, p.XMLHash)
		require.NotEmpty(t, p.Diagnosis.Recommendations)
		require.GreaterOrEqual(t, len(p.Diagnosis.Steps), 5)
	})

	t.Run("bookmark lookup inside deadlock-list wrapper", func(t *testing.T) {
		p, err := ParseXML(bookmarkLookupXML)
		require.NoError(t, err)
		require.Equal(t, PatternBookmarkLookup, p.Diagnosis.Pattern)
		require.Equal(t, "88", p.VictimSPID)
		require.Contains(t, p.Diagnosis.Summary, "IX_Orders_Status")
		require.Contains(t, p.Diagnosis.Summary, "PK_Orders")
	})

	t.Run("lock upgrade conversion", func(t *testing.T) {
		p, err := ParseXML(lockUpgradeXML)
		require.NoError(t, err)
		require.Equal(t, PatternLockUpgrade, p.Diagnosis.Pattern)
		require.Equal(t, "55", p.VictimSPID)
		require.Equal(t, []string{"BillingDB"}, p.Databases)
		require.Equal(t, "convert", p.Resources[0].Waiters[0].RequestType)
	})

	t.Run("parallelism exchange", func(t *testing.T) {
		p, err := ParseXML(exchangeXML)
		require.NoError(t, err)
		require.Equal(t, PatternParallelExchange, p.Diagnosis.Pattern)
		require.Equal(t, "exchangeEvent", p.Resources[0].Type)
	})

	t.Run("empty xml returns error", func(t *testing.T) {
		_, err := ParseXML("   ")
		require.Error(t, err)
	})
}
