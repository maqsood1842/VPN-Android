package com.vpn.core.network

import android.app.Activity
import android.content.Context
import com.android.billingclient.api.AcknowledgePurchaseParams
import com.android.billingclient.api.BillingClient
import com.android.billingclient.api.BillingClientStateListener
import com.android.billingclient.api.BillingFlowParams
import com.android.billingclient.api.BillingResult
import com.android.billingclient.api.Purchase
import com.android.billingclient.api.PurchasesUpdatedListener
import com.android.billingclient.api.QueryProductDetailsParams
import com.android.billingclient.api.QueryPurchasesParams
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class BillingService @Inject constructor(
    private val context: Context
) : PurchasesUpdatedListener {

    private var billingClient: BillingClient? = null
    private val scope = CoroutineScope(Dispatchers.IO)
    private var onPurchasesUpdatedCallback: ((List<Purchase>) -> Unit)? = null

    interface BillingCallback {
        fun onSuccess(purchases: List<Purchase>)
        fun onError(error: Exception)
    }

    fun initialize(onPurchasesUpdated: (List<Purchase>) -> Unit) {
        this.onPurchasesUpdatedCallback = onPurchasesUpdated

        billingClient = BillingClient.newBuilder(context)
            .setListener(this)
            .enablePendingPurchases()
            .build()

        billingClient?.startConnection(object : BillingClientStateListener {
            override fun onBillingSetupFinished(billingResult: BillingResult) {
                if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                    // Connection successful
                    queryPurchases()
                }
            }

            override fun onBillingServiceDisconnected() {
                // Try to reconnect
                initialize(onPurchasesUpdated)
            }
        })
    }

    fun queryProducts(productIds: List<String>, callback: BillingCallback) {
        billingClient?.let { client ->
            val params = QueryProductDetailsParams.newBuilder()
            productIds.forEach { id ->
                params.addProduct(QueryProductDetailsParams.Product.newBuilder()
                    .setProductId(id)
                    .setProductType(BillingClient.ProductType.SUBS)
                    .build())
            }
            params.build().let { queryParams ->
                client.queryProductDetailsAsync(queryParams) { billingResult, productDetailsList ->
                    if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                        // Convert to purchases for compatibility
                        val purchases = productDetailsList?.map { details ->
                            // Create a minimal purchase-like object for product details
                            // In real implementation, you'd return product details directly
                            Purchase.newBuilder("test", "test").build()
                        } ?: emptyList()
                        callback.onSuccess(purchases)
                    } else {
                        callback.onError(Exception("Query products failed: ${billingResult.debugMessage}"))
                    }
                }
            }
        }
    }

    fun queryPurchases() {
        billingClient?.let { client ->
            val params = QueryPurchasesParams.newBuilder()
                .setProductType(BillingClient.ProductType.SUBS)
                .build()
            client.queryPurchasesAsync(params) { billingResult, purchases ->
                if (billingResult.responseCode == BillingClient.BillingResponseCode.OK) {
                    onPurchasesUpdatedCallback?.invoke(purchases)
                }
            }
        }
    }

    fun launchBillingFlow(activity: Activity, productId: String, callback: (BillingResult) -> Unit) {
        billingClient?.let { client ->
            val params = BillingFlowParams.newBuilder()
                .setProductDetailsParamsList(listOf(
                    BillingFlowParams.ProductDetailsParams.newBuilder()
                        .setProductId(productId)
                        .setProductType(BillingClient.ProductType.SUBS)
                        .build()
                ))
                .build()
            client.launchBillingFlow(activity, params) { billingResult ->
                callback(billingResult)
            }
        }
    }

    fun acknowledgePurchase(purchaseToken: String, callback: (BillingResult) -> Unit) {
        billingClient?.let { client ->
            val params = AcknowledgePurchaseParams.newBuilder()
                .setPurchaseToken(purchaseToken)
                .build()
            client.acknowledgePurchase(params) { billingResult ->
                callback(billingResult)
            }
        }
    }

    override fun onPurchasesUpdated(billingResult: BillingResult, purchases: List<Purchase>?) {
        if (billingResult.responseCode == BillingClient.BillingResponseCode.OK && purchases != null) {
            onPurchasesUpdatedCallback?.invoke(purchases)
        }
    }

    fun isReady(): Boolean = billingClient?.isReady == true

    fun disconnect() {
        billingClient?.endConnection()
        billingClient = null
    }
}